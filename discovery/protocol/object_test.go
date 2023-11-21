package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBadStorObject(t *testing.T) {

	ep := Endpoint{
		Services: &Service{
			Name: "x",
			Uuid: "3",
		},
	}

	// Key() function without fatal
	_, err := uuid.Parse(ep.Services.Uuid)
	if err == nil {
		t.Errorf("Should be invalid uuid\n")
	}

	t.Logf("expected error: %v\n", err)

}

func TestGoodStorObject(t *testing.T) {

	ep := Endpoint{
		Services: &Service{
			Name: "x",
			Uuid: uuid.New().String(),
		},
	}

	ep.SetVersion(3)
	assert.Equal(t, ep.Version, int64(3), "version should be equal to 3")

	assert.Equal(t, ep.Key(), fmt.Sprintf("%s/%s", EndpointPrefix, ep.Services.Uuid), "keys should be equal")
}
