/*
 * This file is a copy of https://gitlab.com/mergetb/tech/cogs/-/blob/master/pkg/common/etcd.go
 */

package pkg

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	etcdConfig     *ServiceConfig
	MaxMessageSize = 2 * 1024 * 1024
)

// SetEtcdConfig sets the global etcd configuration settings
func SetEtcdConfig(cfg *ServiceConfig) {
	if cfg == nil {
		log.Fatal("etcd service config cannot be nil")
	}
	etcdConfig = cfg
}

// EtcdConnect Try to get a etcd client- assumption EtcdClient is async until used
func EtcdConnect() (*clientv3.Client, error) {
	log.Trace("connecting to etcd...")

	etcd, err := EtcdClient()
	if err == nil {
		return etcd, nil
	}
	return nil, fmt.Errorf("%v: failed to connect to etcd\n", err)
}

// EnsureEtcd Make sure we always have an etcd connection
func EnsureEtcd(etcdp **clientv3.Client) error {

	// ensure we have a usable etcd connection
	var err error
	if *etcdp == nil {
		log.Debugf("etcd connection nil - connecting")
		*etcdp, err = EtcdConnect()
		if err != nil {
			return err
		}
	}

	// experimental: https://github.com/grpc/grpc-go/pull/1430
	state := (**etcdp).ActiveConnection().GetState()

	//https://github.com/grpc/grpc/blob/master/doc/connectivity-semantics-and-api.md
	if state != connectivity.Ready && state != connectivity.Idle {

		log.Warnf("etcd status check error - reconnecting: %v\n", err)

		(*etcdp).Close()
		*etcdp, err = EtcdConnect()
		if err != nil {
			return err
		}

		state := (**etcdp).ActiveConnection().GetState()

		if state != connectivity.Ready && state != connectivity.Idle {
			return fmt.Errorf("%v: etcd status check failed - giving up\n", err)
		}
	}

	return err

}

// WithEtcd executes a function against an etcd client with a managed
// connection lifetime.
func WithEtcd(f func(*clientv3.Client) error) error {

	cli, err := EtcdConnect()
	if err != nil {
		return err
	}
	defer cli.Close()

	return f(cli)

}

// WithMinEtcd executes a function against an etcd client with a managed
// connection lifetime.
func WithMinEtcd(f func(*clientv3.Client) (interface{}, error)) (interface{}, error) {

	cli, err := EtcdConnect()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	return f(cli)

}

// EtcdClient read the specific configuration files to initiate etcd connection
func EtcdClient() (*clientv3.Client, error) {

	log.Trace("creating new client...")

	cfg := etcdConfig

	var tlsc *tls.Config
	if cfg.TLS != nil {

		log.Trace("etcd tls enabled")

		log.WithFields(log.Fields{
			"cacert": cfg.TLS.Cacert,
			"cert":   cfg.TLS.Cert,
			"key":    cfg.TLS.Key,
		}).Trace("tls config")

		capool := x509.NewCertPool()
		capem, err := ioutil.ReadFile(cfg.TLS.Cacert)
		if err != nil {
			return nil, fmt.Errorf("%v: failed to read cacert\n", err)
		}
		ok := capool.AppendCertsFromPEM(capem)
		if !ok {
			return nil, fmt.Errorf("%v: capem is not ok", err)
		}

		cert, err := tls.LoadX509KeyPair(
			cfg.TLS.Cert,
			cfg.TLS.Key,
		)
		if err != nil {
			return nil, fmt.Errorf("%v: failed to load cert/key pair\n", err)
		}

		tlsc = &tls.Config{
			RootCAs:      capool,
			Certificates: []tls.Certificate{cert},
		}
	} else {
		log.Trace("etcd tls disabled")
	}

	connstr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	log.WithFields(log.Fields{
		"connstr": connstr,
	}).Trace("etcd connection string")

	// The issue here with etcd is that the connection to the database is
	// sticking around for 2MSL (2 minutes), so we will run into issues with
	// max number of connections.
	// So we will pass a dialoption, with a dialler, that overwrites the
	// standard tcp connection with SO_LINGER.  Setting to 0 deletes immediately,
	// > 1 is seconds, < 0 is backgrounded. non-zero leaves time_wait for 2MSL
	//TODO: There should be a better way of tracking down why the connection isnt
	//closing correctly.
	f := func(ctx context.Context, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Deadline: time.Now().Add(1 * time.Minute),
		}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		//http://www.serverframework.com/asynchronousevents/2011/01/
		//time-wait-and-its-design-implications-for-protocols-and-
		//scalable-servers.html
		//conn.(*net.TCPConn).SetLinger(0)
		err = conn.(*net.TCPConn).SetKeepAlive(true)
		if err != nil {
			log.Warnf("failure to keepalive: %v\n", err)
		}
		err = conn.(*net.TCPConn).SetKeepAlivePeriod(1 * time.Second)
		if err != nil {
			log.Warnf("failure to set keepalive period: %v\n", err)
		}
		return conn, err
	}

	opts := []grpc.DialOption{
		grpc.WithContextDialer(f),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxMessageSize),
			grpc.MaxCallSendMsgSize(MaxMessageSize),
		),
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{connstr},
		DialTimeout:          3 * time.Second,
		DialKeepAliveTime:    -1 * time.Second,
		DialKeepAliveTimeout: -1 * time.Second,
		TLS:                  tlsc,
		DialOptions:          opts,
		MaxCallSendMsgSize:   MaxMessageSize,
		MaxCallRecvMsgSize:   MaxMessageSize,
	})

	log.Trace("client created")
	return cli, err

}
