# Running

```
./run.sh
./update_hosts.sh
./format_disks.sh
./configure_hosts.sh

#May be required- to mount our nfs mounts
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i .rvn/ansible-hosts retry-mount.yml

sudo rvn configure

./setup_tests.sh
```

At this state we should have a microk8s setup with helm installed and our pod has been created in the `orchestrator` namespace:

```
rvn@orchestrator:~$ sudo kubectl -n orch get pods -o=wide
NAME                            READY   STATUS    RESTARTS   AGE   IP             NODE           NOMINATED NODE   READINESS GATES
orchestrator-54c866c679-4qjbf   5/5     Running   0          32m   10.1.142.107   orchestrator   <none>           <none>
```

Adding the values directly into etcd:
```
rvn@orchestrator:~$ sudo kubectl -n orch exec -it $POD -c mock-discovery /bin/bash 
root@orchestrator-69b5987cd9-xh9qk:/# curl http://localhost:15015/mock
```

etcd command
```
ETCDCTL_API=3 etcdctl get --prefix ""
ETCDCTL_API=3 etcdctl get --keys-only --prefix ""
```

## Modifying

I either re-run `./setup_tests.sh` to tear down all the resources and rebuild from scratch, or I run:

```
POD=$(sudo kubectl -n orch get pods | grep orchestrator | cut -d " " -f 1)
sudo kubectl -n orch delete pod ${POD
```

Since we manage the pod through helm, deleting the pod causes k8s to re-deploy the resouces.

# Structure

Currently we spin up a pod from helm in the "orchestrator" namespace.

There is a single pod with at the moment 5 containers:

* etcd
* discovery-api
* discovery-scanner
* mock-discovery
* inventory-api

Currently etcd is emphemeral with the pod, a production implementation will use a PV backing to maintain state across pod modifications.

All the services use etcd, in the same etcd namespace, later we will change this to use grpc-proxies which limit each services access.  Access to etcd is plumbed to a configuration file passed through as a config-map, so changing the default values to ones assigned to each service's grpc proxy will limit key namespace access.

The discovery-api handles "Endpoint" accesses, that is adding, modifying, deleting, endpoints.  An endpoint in this architecture refers to orchestrators which have a set of resources.

The discovery-scanner uses the endpoints managed by the api and stored in etcd to probe the resources of each endpoint (core).  At the moment, to our knowledge no core readily exposes the resources, so we've mocked up (mock-discovery) a server that returns a resource struct.  That resource-struct is then used by the cbs-solver to find solutions for constraint-based requests.


For now, to do a simple test.  All the pods come up, there isnt anything happening.

## Add data to discovery service

So first we need to add an endpoint for the discovery scanner to find:

```
POD=$(sudo kubectl -n orch get pods | grep orchestrator | cut -d " " -f 1)
rvn@orchestrator:~$ sudo kubectl -n orch exec -it $POD -c discovery-api /bin/bash
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [POD] -- [COMMAND] instead.
root@orchestrator-54c866c679-57k2b:/# dctl create disc /data/discovery/pkg/test_service_config_1.yml
```

The file being added is mounted from in the pod, and is a test file that specifies a mocked host (which if mock-discovery is running - is the same endpoint).

This will cause the scanner which is constantly scanning the database to see a new endpoint and scan it.  To which it find the `/resource` endpoint which has a resourceItem attached to it.  From here the scanner checks if the item is already in the inventory, if it is, it does nothing with it, if the data has changed it updates inventory.

## Create a network map from scanned resources

```
POD=$(sudo kubectl -n orch get pods | grep orchestrator | cut -d " " -f 1)
rvn@orchestrator:~$ POD=$(sudo kubectl -n orch get pods | grep orchestrator | cut -d " " -f 1)
rvn@orchestrator:~$ sudo kubectl -n orch exec $POD -c network -- /usr/bin/snctl create
sent request
rvn@orchestrator:~$ sudo kubectl -n orch exec $POD -c network -- /usr/bin/snctl show
digraph "" {
....
}
```


## Sending CBS request

standalone method:
```
curl -X POST -H "Content-Type: application/json" -d @./pkg/mockcbs.request http://localhost:15030/cbs | jq .
```

integrated method:

```
```
