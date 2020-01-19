package layer // import "github.com/docker/docker/layer"

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

/* DigestSHA256EmptyTar 的 sha256 digest 的值可由下面的方法实验得到：
[root@MiWiFi-R4AC-srv ~]# dd bs=1024 count=1 if=/dev/zero | sha256sum
1+0 records in
1+0 records out
1024 bytes (1.0 kB) copied, 0.0002037 s, 5.0 MB/s
5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef  -

使用 1204 zero bytes 作为 empty tar 是由 tar 格式决定的，参考：
https://www.gnu.org/software/tar/manual/html_node/Standard.html

	Physically, an archive consists of a series of file entries terminated 
	by an end-of-archive entry, which consists of two 512 blocks of zero bytes. 
*/

// DigestSHA256EmptyTar is the canonical sha256 digest of empty tar file -
// (1024 NULL bytes)
const DigestSHA256EmptyTar = DiffID("sha256:5f70bf18a086007016e948b04aed3b82103a36bea41755b6cddfaf10ace3c6ef")

type emptyLayer struct{}

// EmptyLayer is a layer that corresponds to empty tar.
var EmptyLayer = &emptyLayer{}

func (el *emptyLayer) TarStream() (io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buf)
	tarWriter.Close()
	return ioutil.NopCloser(buf), nil
}

func (el *emptyLayer) TarStreamFrom(p ChainID) (io.ReadCloser, error) {
	if p == "" {
		return el.TarStream()
	}
	return nil, fmt.Errorf("can't get parent tar stream of an empty layer")
}

func (el *emptyLayer) ChainID() ChainID {
	return ChainID(DigestSHA256EmptyTar)
}

func (el *emptyLayer) DiffID() DiffID {
	return DigestSHA256EmptyTar
}

func (el *emptyLayer) Parent() Layer {
	return nil
}

func (el *emptyLayer) Size() (size int64, err error) {
	return 0, nil
}

func (el *emptyLayer) DiffSize() (size int64, err error) {
	return 0, nil
}

func (el *emptyLayer) Metadata() (map[string]string, error) {
	return make(map[string]string), nil
}

// IsEmpty returns true if the layer is an EmptyLayer
func IsEmpty(diffID DiffID) bool {
	return diffID == DigestSHA256EmptyTar
}
