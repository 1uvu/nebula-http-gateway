package v2_6

import (
	"github.com/facebook/fbthrift/thrift/lib/go/thrift"

	nerrors "github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/errors"
	nthrift "github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/internal/thrift/v2_6"
	"github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/internal/thrift/v2_6/meta"
	"github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/types"
)

var (
	_ types.MetaClientDriver = (*defaultMetaClient)(nil)
)

type (
	defaultMetaClient struct {
		meta *meta.MetaServiceClient
	}
)

func newMetaClient(transport thrift.Transport, pf thrift.ProtocolFactory) types.MetaClientDriver {
	return &defaultMetaClient{
		meta: meta.NewMetaServiceClientFactory(transport, pf),
	}
}

func (c *defaultMetaClient) Open() error {
	return c.meta.Open()
}

func (c *defaultMetaClient) VerifyClientVersion() error {
	req := meta.NewVerifyClientVersionReq()
	resp, err := c.meta.VerifyClientVersion(req)
	if err != nil {
		return err
	}
	return codeErrorIfHappened(resp.Code, resp.ErrorMsg)
}

func (c *defaultMetaClient) Close() error {
	if c.meta != nil {
		if err := c.meta.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *defaultMetaClient) AddHosts(endpoints []string) error {
	return nerrors.ErrUnsupported
}

func (c *defaultMetaClient) DropHosts(endpoints []string) error {
	return nerrors.ErrUnsupported
}

func (c *defaultMetaClient) ListSpaces() (types.ListSpacesResponse, error) {
	req := meta.NewListSpacesReq()

	resp, err := c.meta.ListSpaces(req)
	if err != nil {
		return nil, err
	}

	return newListSpacesResponseWrapper(resp.Spaces), codeErrorIfHappened(resp.Code, []byte(nthrift.ErrorCodeToName[resp.Code]))
}

func (c *defaultMetaClient) SubmitJobBalance(space string, endpoints ...string) error {
	return nerrors.ErrUnsupported
}
