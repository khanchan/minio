package minioapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	mstorage "github.com/minio-io/minio/pkg/storage"
)

// Write Common Header helpers
func writeCommonHeaders(w http.ResponseWriter, acceptsType string) {
	w.Header().Set("Server", "Minio")
	w.Header().Set("Content-Type", acceptsType)
}

func writeErrorResponse(w http.ResponseWriter, response interface{}, acceptsType contentType) []byte {
	var bytesBuffer bytes.Buffer
	var encoder encoder
	// write common headers
	writeCommonHeaders(w, getContentString(acceptsType))
	switch acceptsType {
	case xmlType:
		encoder = xml.NewEncoder(&bytesBuffer)
	case jsonType:
		encoder = json.NewEncoder(&bytesBuffer)
	}
	encoder.Encode(response)
	return bytesBuffer.Bytes()
}

// Write Object Header helper
func writeObjectHeaders(w http.ResponseWriter, metadata mstorage.ObjectMetadata) {
	lastModified := metadata.Created.Format(time.RFC1123)
	// write common headers
	writeCommonHeaders(w, metadata.ContentType)
	w.Header().Set("ETag", metadata.ETag)
	w.Header().Set("Last-Modified", lastModified)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("Connection", "close")
}

func writeObjectHeadersAndResponse(w http.ResponseWriter, response interface{}, acceptsType contentType) []byte {
	var bytesBuffer bytes.Buffer
	var encoder encoder
	// write common headers
	writeCommonHeaders(w, getContentString(acceptsType))
	switch acceptsType {
	case xmlType:
		encoder = xml.NewEncoder(&bytesBuffer)
	case jsonType:
		encoder = json.NewEncoder(&bytesBuffer)
	}

	w.Header().Set("Connection", "close")
	encoder.Encode(response)
	return bytesBuffer.Bytes()
}