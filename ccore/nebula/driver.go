package nebula

import (
	"sync"

	"github.com/facebook/fbthrift/thrift/lib/go/thrift"
	nerrors "github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/errors"
	_ "github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/internal/driver/v2_5"
	_ "github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/internal/driver/v2_6"
	_ "github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/internal/driver/v3_0"
	"github.com/vesoft-inc/nebula-http-gateway/ccore/nebula/types"
)

type (
	driverGraph struct {
		types.GraphClientDriver
		connection *connectionMu
		username   string
		password   string
		sessionId  int64
	}

	driverMeta struct {
		types.MetaClientDriver
		connection *connectionMu
	}

	driverStorageAdmin struct {
		types.StorageAdminClientDriver
		connection *connectionMu
	}

	connectionMu struct {
		o         *socketOptions
		mu        sync.Mutex
		endpoints []string
	}

	driverFactory struct {
		types.FactoryDriver
	}
)

func newDriverGraph(endpoints []string, username, password string, o *socketOptions) *driverGraph {
	return &driverGraph{
		connection: newConnectionMu(endpoints, o),
		username:   username,
		password:   password,
	}
}

func newDriverMeta(endpoints []string, o *socketOptions) *driverMeta {
	return &driverMeta{
		connection: newConnectionMu(endpoints, o),
	}
}

func newDriverStorageAdmin(endpoints []string, o *socketOptions) *driverStorageAdmin {
	return &driverStorageAdmin{
		connection: newConnectionMu(endpoints, o),
	}
}

func newConnectionMu(endpoints []string, o *socketOptions) *connectionMu {
	return &connectionMu{
		o:         o,
		endpoints: endpoints,
	}
}

func newDriverFactory() *driverFactory {
	return &driverFactory{}
}

func (d *driverGraph) open(driver types.Driver) error {
	if d.GraphClientDriver != nil {
		return nil
	}

	transport, pf, err := d.connection.connect()
	if err != nil {
		return err
	}

	graphClientDriver := driver.NewGraphClientDriver(transport, pf)

	if err = graphClientDriver.Open(); err != nil {
		return err
	}

	if err = graphClientDriver.VerifyClientVersion(); err != nil {
		return err
	}

	resp, err := graphClientDriver.Authenticate(d.username, d.password)
	if err != nil {
		return err
	}

	sessionId := resp.SessionID()
	if sessionId == nil {
		panic("sessionId can not be nil after authenticate")
	}
	d.sessionId = *sessionId

	d.GraphClientDriver = graphClientDriver
	return nil
}

func (d *driverGraph) close() error {
	if d.GraphClientDriver != nil {
		if err := d.GraphClientDriver.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (d *driverMeta) open(driver types.Driver) error {
	if d.MetaClientDriver != nil {
		return nil
	}

	transport, pf, err := d.connection.connect()
	if err != nil {
		return err
	}

	metaClientDriver := driver.NewMetaClientDriver(transport, pf)

	if err = metaClientDriver.Open(); err != nil {
		return err
	}

	if err = metaClientDriver.VerifyClientVersion(); err != nil {
		return err
	}

	d.MetaClientDriver = metaClientDriver
	return nil
}

func (d *driverMeta) close() error {
	if d.MetaClientDriver != nil {
		if err := d.MetaClientDriver.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (d *driverStorageAdmin) open(driver types.Driver) error {
	if d.StorageAdminClientDriver != nil {
		return nil
	}

	transport, pf, err := d.connection.connect()
	if err != nil {
		return err
	}

	storageAdminClientDriver := driver.NewStorageClientDriver(transport, pf)

	if err = storageAdminClientDriver.Open(); err != nil {
		return err
	}

	d.StorageAdminClientDriver = storageAdminClientDriver
	return nil
}

func (d *driverStorageAdmin) close() error {
	if d.StorageAdminClientDriver != nil {
		if err := d.StorageAdminClientDriver.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *connectionMu) connect() (thrift.Transport, thrift.ProtocolFactory, error) {
	// TODO: automatically open until success, only the first endpoints is supported now.
	if len(c.endpoints) == 0 {
		return nil, nil, nerrors.ErrNoEndpoints
	}
	return c.buildThriftTransport(c.endpoints[0])
}

func (c *connectionMu) buildThriftTransport(endpoint string) (thrift.Transport, thrift.ProtocolFactory, error) {
	transport, err := func() (thrift.Transport, error) {
		if c.o.tlsConfig == nil {
			return thrift.NewSocket(thrift.SocketTimeout(c.o.timeout), thrift.SocketAddr(endpoint))
		}
		return thrift.NewSSLSocketTimeout(endpoint, c.o.tlsConfig, c.o.timeout)
	}()
	if err != nil {
		return nil, nil, err
	}

	bufferedTranFactory := thrift.NewBufferedTransportFactory(c.o.bufferSize)
	transport = thrift.NewFramedTransportMaxLength(bufferedTranFactory.GetTransport(transport), c.o.frameMaxLength)
	pf := thrift.NewBinaryProtocolFactoryDefault()

	return transport, pf, nil
}
