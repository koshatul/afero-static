package afstatic

type CompressionType string

const (
	NoCompression   CompressionType = "none"
	DeflateCompress CompressionType = "deflate"
	GZipCompress    CompressionType = "gzip"
	LzwCompress     CompressionType = "lzw"
	SnappyCompress  CompressionType = "snappy"
	ZlibCompress    CompressionType = "zlib"
)

var compressionList = map[string]CompressionType{
	string(NoCompression):   NoCompression,
	string(DeflateCompress): DeflateCompress,
	string(GZipCompress):    GZipCompress,
	string(LzwCompress):     LzwCompress,
	string(SnappyCompress):  SnappyCompress,
	string(ZlibCompress):    ZlibCompress,
}
