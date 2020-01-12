package layer // import "github.com/docker/docker/layer"

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/docker/pkg/ioutils"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	stringIDRegexp      = regexp.MustCompile(`^[a-f0-9]{64}(-init)?$`)
	supportedAlgorithms = []digest.Algorithm{
		digest.SHA256,
		// digest.SHA384, // Currently not used
		// digest.SHA512, // Currently not used
	}
)

/* fileMetadataStore 是一个存储“层”的元数据的 store，其中 root 在 daemon 的初始化过程中（daemon.NewDaemon）被设置
为 /var/lib/docker/image/<driver>/layerdb，目录结构如下：

	/var/lib/docker/image/<driver>/layerdb
	  - mounts
	    - <containerId>
	      - init-id
	      - mount-id
	      - parent
	    - ...
	  - sha256
	    - <chainId>
	      - cache-id
	      - diff
	      - parent
	      - size
	      - tar-split.json.gz
	    - ...
	  - tmp

* 其中，mounts 中存储的是各个容器的 init 层（只读）和读写层的元数据，sha256 中存储的是镜像层的元数据，tmp 是一个临时目录。
  对于 mounts/<containerId>中的文件：
	+ mount-id	保存了读写层的 id，这个 id 是在容器创建时使用 github.com/docker/docker/pkg/stringid 生产的一个随机的长度为 64 的字符串；
	+ init-id	保存了读写层的 id 加后缀 -init 后的字符串，init 层存在的必要性可参考
	  https://stackoverflow.com/questions/33697774/init-layer-in-docker、http://www.projectatomic.io/docs/filesystems/；
	+ parent	保存了读写层父 layer 的 chainId，注意，这个 chainId 对应的不是 init 层；
  对于 sha256/<chainId> 中的文件：
	+ cache-id	保存了 layer 数据存放位置的相关信息，例如：假如 driver 为 aufs，则相应 layer 数据存储在 /var/lib/docker/aufs/diff/<cache-id>，
	  假如 driver 为 overlay2，那么层数据存放在 /var/lib/docker/overlay2/<cache-id>，假如 driver 为 devicemapper，则层数据存放的 device-id 会
	  存储在 /var/lib/docker/devicemapper/metadata/<cache-id>中，cache-id 用于定位层数据；
	+ diff-id	保存了层的 diffId，diffId 是层 tar 格式的 hash 值，用于校验层的完整性；
	+ parent	保存了该层父 layer 的 chainId；
	+ size		保存了层的大小；
	+ tar-split.json.gz 存放的是 layer 的 json 信息；
*/
type fileMetadataStore struct {
	root string
}

type fileMetadataTransaction struct {
	store *fileMetadataStore
	ws    *ioutils.AtomicWriteSet
}

// newFSMetadataStore returns an instance of a metadata store
// which is backed by files on disk using the provided root
// as the root of metadata files.
func newFSMetadataStore(root string) (*fileMetadataStore, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	return &fileMetadataStore{
		root: root,
	}, nil
}

// 返回结果是目录 /var/lib/docker/image/<driver>/layerdb/sha256/<chainId>，包含五个文件：
// cache-id、diff、parent、size、tar-split.json.gz
func (fms *fileMetadataStore) getLayerDirectory(layer ChainID) string {
	dgst := digest.Digest(layer)
	return filepath.Join(fms.root, string(dgst.Algorithm()), dgst.Hex())
}

// 获取 /var/lib/docker/image/<driver>/layerdb/sha256/<chainId> 下的某个文件
func (fms *fileMetadataStore) getLayerFilename(layer ChainID, filename string) string {
	return filepath.Join(fms.getLayerDirectory(layer), filename)
}

// 返回的结果是目录 /var/lib/docker/image/<driver>/layerdb/mounts/<containerId>，其中包含三个文件：
// init-id、mount-id、parent
func (fms *fileMetadataStore) getMountDirectory(mount string) string {
	return filepath.Join(fms.root, "mounts", mount)
}

// 获取 /var/lib/docker/image/<driver>/layerdb/mounts/<containerId> 下的某个文件
func (fms *fileMetadataStore) getMountFilename(mount, filename string) string {
	return filepath.Join(fms.getMountDirectory(mount), filename)
}

func (fms *fileMetadataStore) StartTransaction() (*fileMetadataTransaction, error) {
	tmpDir := filepath.Join(fms.root, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, err
	}
	ws, err := ioutils.NewAtomicWriteSet(tmpDir)
	if err != nil {
		return nil, err
	}

	return &fileMetadataTransaction{
		store: fms,
		ws:    ws,
	}, nil
}

// 设置 /var/lib/docker/image/<driver>/layerdb/tmp/size
func (fm *fileMetadataTransaction) SetSize(size int64) error {
	content := fmt.Sprintf("%d", size)
	return fm.ws.WriteFile("size", []byte(content), 0644)
}

// 设置 /var/lib/docker/image/<driver>/layerdb/tmp/parent
func (fm *fileMetadataTransaction) SetParent(parent ChainID) error {
	return fm.ws.WriteFile("parent", []byte(digest.Digest(parent).String()), 0644)
}

// 设置 /var/lib/docker/image/<driver>/layerdb/tmp/diff
func (fm *fileMetadataTransaction) SetDiffID(diff DiffID) error {
	return fm.ws.WriteFile("diff", []byte(digest.Digest(diff).String()), 0644)
}

// 设置 /var/lib/docker/image/<driver>/layerdb/tmp/cache-id
func (fm *fileMetadataTransaction) SetCacheID(cacheID string) error {
	return fm.ws.WriteFile("cache-id", []byte(cacheID), 0644)
}

func (fm *fileMetadataTransaction) SetDescriptor(ref distribution.Descriptor) error {
	jsonRef, err := json.Marshal(ref)
	if err != nil {
		return err
	}
	return fm.ws.WriteFile("descriptor.json", jsonRef, 0644)
}

func (fm *fileMetadataTransaction) TarSplitWriter(compressInput bool) (io.WriteCloser, error) {
	f, err := fm.ws.FileWriter("tar-split.json.gz", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	var wc io.WriteCloser
	if compressInput {
		wc = gzip.NewWriter(f)
	} else {
		wc = f
	}

	return ioutils.NewWriteCloserWrapper(wc, func() error {
		wc.Close()
		return f.Close()
	}), nil
}

func (fm *fileMetadataTransaction) Commit(layer ChainID) error {
	finalDir := fm.store.getLayerDirectory(layer)
	if err := os.MkdirAll(filepath.Dir(finalDir), 0755); err != nil {
		return err
	}

	return fm.ws.Commit(finalDir)
}

func (fm *fileMetadataTransaction) Cancel() error {
	return fm.ws.Cancel()
}

func (fm *fileMetadataTransaction) String() string {
	return fm.ws.String()
}

// 读取 /var/lib/docker/image/<driver>/layerdb/sha256/<chainId>/size
func (fms *fileMetadataStore) GetSize(layer ChainID) (int64, error) {
	content, err := ioutil.ReadFile(fms.getLayerFilename(layer, "size"))
	if err != nil {
		return 0, err
	}

	size, err := strconv.ParseInt(string(content), 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// 读取 /var/lib/docker/image/<driver>/layerdb/sha256/<chainId>/parent
func (fms *fileMetadataStore) GetParent(layer ChainID) (ChainID, error) {
	content, err := ioutil.ReadFile(fms.getLayerFilename(layer, "parent"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	dgst, err := digest.Parse(strings.TrimSpace(string(content)))
	if err != nil {
		return "", err
	}

	return ChainID(dgst), nil
}

// 读取 /var/lib/docker/image/<driver>/layerdb/sha256/<chainId>/diff
func (fms *fileMetadataStore) GetDiffID(layer ChainID) (DiffID, error) {
	content, err := ioutil.ReadFile(fms.getLayerFilename(layer, "diff"))
	if err != nil {
		return "", err
	}

	dgst, err := digest.Parse(strings.TrimSpace(string(content)))
	if err != nil {
		return "", err
	}

	return DiffID(dgst), nil
}

// 读取 /var/lib/docker/image/<driver>/layerdb/sha256/<chainId>/cache-id
func (fms *fileMetadataStore) GetCacheID(layer ChainID) (string, error) {
	contentBytes, err := ioutil.ReadFile(fms.getLayerFilename(layer, "cache-id"))
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))

	if content == "" {
		return "", errors.Errorf("invalid cache id value")
	}

	return content, nil
}

func (fms *fileMetadataStore) GetDescriptor(layer ChainID) (distribution.Descriptor, error) {
	content, err := ioutil.ReadFile(fms.getLayerFilename(layer, "descriptor.json"))
	if err != nil {
		if os.IsNotExist(err) {
			// only return empty descriptor to represent what is stored
			return distribution.Descriptor{}, nil
		}
		return distribution.Descriptor{}, err
	}

	var ref distribution.Descriptor
	err = json.Unmarshal(content, &ref)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	return ref, err
}

func (fms *fileMetadataStore) TarSplitReader(layer ChainID) (io.ReadCloser, error) {
	fz, err := os.Open(fms.getLayerFilename(layer, "tar-split.json.gz"))
	if err != nil {
		return nil, err
	}
	f, err := gzip.NewReader(fz)
	if err != nil {
		fz.Close()
		return nil, err
	}

	return ioutils.NewReadCloserWrapper(f, func() error {
		f.Close()
		return fz.Close()
	}), nil
}

// 设置 /var/lib/docker/image/overlay2/layerdb/mounts/<containerId>/mount-id
func (fms *fileMetadataStore) SetMountID(mount string, mountID string) error {
	if err := os.MkdirAll(fms.getMountDirectory(mount), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(fms.getMountFilename(mount, "mount-id"), []byte(mountID), 0644)
}

// 设置 /var/lib/docker/image/overlay2/layerdb/mounts/<containerId>/init-id
func (fms *fileMetadataStore) SetInitID(mount string, init string) error {
	if err := os.MkdirAll(fms.getMountDirectory(mount), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(fms.getMountFilename(mount, "init-id"), []byte(init), 0644)
}

// 设置 /var/lib/docker/image/overlay2/layerdb/mounts/<containerId>/parent
func (fms *fileMetadataStore) SetMountParent(mount string, parent ChainID) error {
	if err := os.MkdirAll(fms.getMountDirectory(mount), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(fms.getMountFilename(mount, "parent"), []byte(digest.Digest(parent).String()), 0644)
}

// 获取 /var/lib/docker/image/overlay2/layerdb/mounts/<containerId>/mount-id
func (fms *fileMetadataStore) GetMountID(mount string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fms.getMountFilename(mount, "mount-id"))
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))

	if !stringIDRegexp.MatchString(content) {
		return "", errors.New("invalid mount id value")
	}

	return content, nil
}

// 获取 /var/lib/docker/image/overlay2/layerdb/mounts/<containerId>/init-id
func (fms *fileMetadataStore) GetInitID(mount string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fms.getMountFilename(mount, "init-id"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))

	if !stringIDRegexp.MatchString(content) {
		return "", errors.New("invalid init id value")
	}

	return content, nil
}

// 获取 /var/lib/docker/image/overlay2/layerdb/mounts/<containerId>/parent
func (fms *fileMetadataStore) GetMountParent(mount string) (ChainID, error) {
	content, err := ioutil.ReadFile(fms.getMountFilename(mount, "parent"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	dgst, err := digest.Parse(strings.TrimSpace(string(content)))
	if err != nil {
		return "", err
	}

	return ChainID(dgst), nil
}

func (fms *fileMetadataStore) getOrphan() ([]roLayer, error) {
	var orphanLayers []roLayer
	for _, algorithm := range supportedAlgorithms {
		fileInfos, err := ioutil.ReadDir(filepath.Join(fms.root, string(algorithm)))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, fi := range fileInfos {
			if !fi.IsDir() || !strings.HasSuffix(fi.Name(), "-removing") {
				continue
			}
			// At this stage, fi.Name value looks like <digest>-<random>-removing
			// Split on '-' to get the digest value.
			nameSplit := strings.Split(fi.Name(), "-")
			dgst := digest.NewDigestFromEncoded(algorithm, nameSplit[0])
			if err := dgst.Validate(); err != nil {
				logrus.WithError(err).WithField("digest", string(algorithm)+":"+nameSplit[0]).Debug("ignoring invalid digest")
				continue
			}

			chainFile := filepath.Join(fms.root, string(algorithm), fi.Name(), "cache-id")
			contentBytes, err := ioutil.ReadFile(chainFile)
			if err != nil {
				if !os.IsNotExist(err) {
					logrus.WithError(err).WithField("digest", dgst).Error("failed to read cache ID")
				}
				continue
			}
			cacheID := strings.TrimSpace(string(contentBytes))
			if cacheID == "" {
				logrus.Error("invalid cache ID")
				continue
			}

			l := &roLayer{
				chainID: ChainID(dgst),
				cacheID: cacheID,
			}
			orphanLayers = append(orphanLayers, *l)
		}
	}

	return orphanLayers, nil
}

func (fms *fileMetadataStore) List() ([]ChainID, []string, error) {
	var ids []ChainID
	for _, algorithm := range supportedAlgorithms {
		fileInfos, err := ioutil.ReadDir(filepath.Join(fms.root, string(algorithm)))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}

		for _, fi := range fileInfos {
			if fi.IsDir() && fi.Name() != "mounts" {
				dgst := digest.NewDigestFromHex(string(algorithm), fi.Name())
				if err := dgst.Validate(); err != nil {
					logrus.Debugf("Ignoring invalid digest %s:%s", algorithm, fi.Name())
				} else {
					ids = append(ids, ChainID(dgst))
				}
			}
		}
	}

	fileInfos, err := ioutil.ReadDir(filepath.Join(fms.root, "mounts"))
	if err != nil {
		if os.IsNotExist(err) {
			return ids, []string{}, nil
		}
		return nil, nil, err
	}

	var mounts []string
	for _, fi := range fileInfos {
		if fi.IsDir() {
			mounts = append(mounts, fi.Name())
		}
	}

	return ids, mounts, nil
}

// Remove layerdb folder if that is marked for removal
func (fms *fileMetadataStore) Remove(layer ChainID, cache string) error {
	dgst := digest.Digest(layer)
	files, err := ioutil.ReadDir(filepath.Join(fms.root, string(dgst.Algorithm())))
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "-removing") || !strings.HasPrefix(f.Name(), dgst.Encoded()) {
			continue
		}

		// Make sure that we only remove layerdb folder which points to
		// requested cacheID
		dir := filepath.Join(fms.root, string(dgst.Algorithm()), f.Name())
		chainFile := filepath.Join(dir, "cache-id")
		contentBytes, err := ioutil.ReadFile(chainFile)
		if err != nil {
			logrus.WithError(err).WithField("file", chainFile).Error("cannot get cache ID")
			continue
		}
		cacheID := strings.TrimSpace(string(contentBytes))
		if cacheID != cache {
			continue
		}
		logrus.Debugf("Removing folder: %s", dir)
		err = os.RemoveAll(dir)
		if err != nil && !os.IsNotExist(err) {
			logrus.WithError(err).WithField("name", f.Name()).Error("cannot remove layer")
			continue
		}
	}
	return nil
}

// 删除 /var/lib/docker/image/<driver>/layerdb/mounts/<containerId>
func (fms *fileMetadataStore) RemoveMount(mount string) error {
	return os.RemoveAll(fms.getMountDirectory(mount))
}
