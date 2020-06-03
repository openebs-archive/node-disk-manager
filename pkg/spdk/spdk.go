/*
Copyright 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// +build linux, cgo

package spdk

/*
#include "stdlib.h"
#include "stdint.h"

#define SPDK_BLOBSTORE_TYPE_LENGTH 16

typedef uint64_t spdk_blob_id;

typedef struct {
        char bstype[SPDK_BLOBSTORE_TYPE_LENGTH];
}spdk_bs_type;

struct spdk_bs_super_block {
        uint8_t         signature[8];
        uint32_t        version;
        uint32_t        length;
        uint32_t        clean;
		spdk_blob_id    super_blob;
		uint32_t        cluster_size;
		uint32_t        used_page_mask_start;
		uint32_t        used_page_mask_len;
		uint32_t        used_cluster_mask_start;
		uint32_t        used_cluster_mask_len;
		uint32_t        md_start;
		uint32_t        md_len;
		spdk_bs_type     bstype;
		uint32_t        used_blobid_mask_start;
		uint32_t        used_blobid_mask_len;
		uint64_t        size;
		uint32_t        io_unit_size;
		uint8_t         reserved[4000];
		uint32_t        crc;
};

char *get_signature(struct spdk_bs_super_block *spdk)
{
	return spdk->signature;
}
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const spdkSignature = "SPDKBLOB"

type DeviceIdentifier struct {
	DevPath string
}

// IsSPDKSignatureExist check is the signature matches the spdk super block signature
func IsSPDKSignatureExist(signature string) bool {
	// need to compare only the first 8 characters of signature
	if len(signature) > 8 {
		signature = signature[0:8]
	}
	return signature == spdkSignature
}

// GetSPDKSuperBlockSignature tries to read spdk super block from a disk and returns the signature
func (di *DeviceIdentifier) GetSPDKSuperBlockSignature() (string, error) {
	var spdk *C.struct_spdk_bs_super_block
	buf := make([]byte, C.sizeof_struct_spdk_bs_super_block)
	f, err := os.Open(di.DevPath)
	defer f.Close()
	if err != nil {
		return "", err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	err = binary.Read(f, binary.BigEndian, buf)
	if err != nil {
		return "", fmt.Errorf("error reading from %s: %v", di.DevPath, err)
	}

	// converting the read bytes to spdk super block struct
	spdk = (*C.struct_spdk_bs_super_block)(C.CBytes(buf))

	var ptr *C.char
	ptr = (*C.char)(C.get_signature(spdk))
	return C.GoString(ptr), nil
}
