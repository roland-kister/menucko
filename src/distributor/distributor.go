package distributor

import "bytes"

type Distributor interface {
	Distribute(content *bytes.Buffer) error
}
