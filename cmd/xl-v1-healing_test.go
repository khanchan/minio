/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"testing"
)

// Tests healing of format XL.
func TestHealFormatXL(t *testing.T) {
	root, err := newTestConfig("us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	defer removeAll(root)

	nDisks := 16
	fsDirs, err := getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err := parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Everything is fine, should return nil
	obj, _, err := initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl := obj.(*xlObjects)
	if err = healFormatXL(xl.storageDisks); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}

	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Disks 0..15 are nil
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	for i := 0; i <= 15; i++ {
		xl.storageDisks[i] = nil
	}

	if err = healFormatXL(xl.storageDisks); err != errXLReadQuorum {
		t.Fatal("Got an unexpected error: ", err)
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// One disk returns Faulty Disk
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	for i := range xl.storageDisks {
		posixDisk, ok := xl.storageDisks[i].(*posix)
		if !ok {
			t.Fatal("storage disk is not *posix type")
		}
		xl.storageDisks[i] = newNaughtyDisk(posixDisk, nil, errDiskFull)
	}
	if err = healFormatXL(xl.storageDisks); err != errXLReadQuorum {
		t.Fatal("Got an unexpected error: ", err)
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// One disk is not found, heal corrupted disks should return nil
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	xl.storageDisks[0] = nil
	if err = healFormatXL(xl.storageDisks); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Remove format.json of all disks
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	for i := 0; i <= 15; i++ {
		if err = xl.storageDisks[i].DeleteFile(".minio.sys", "format.json"); err != nil {
			t.Fatal(err)
		}
	}
	if err = healFormatXL(xl.storageDisks); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupted format json in one disk
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	for i := 0; i <= 15; i++ {
		if err = xl.storageDisks[i].AppendFile(".minio.sys", "format.json", []byte("corrupted data")); err != nil {
			t.Fatal(err)
		}
	}
	if err = healFormatXL(xl.storageDisks); err == nil {
		t.Fatal("Should get a json parsing error, ")
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Remove format.json on 3 disks.
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	for i := 0; i <= 2; i++ {
		if err = xl.storageDisks[i].DeleteFile(".minio.sys", "format.json"); err != nil {
			t.Fatal(err)
		}
	}
	if err = healFormatXL(xl.storageDisks); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// One disk is not found, heal corrupted disks should return nil
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	for i := 0; i <= 2; i++ {
		if err = xl.storageDisks[i].DeleteFile(".minio.sys", "format.json"); err != nil {
			t.Fatal(err)
		}
	}
	posixDisk, ok := xl.storageDisks[3].(*posix)
	if !ok {
		t.Fatal("storage disk is not *posix type")
	}
	xl.storageDisks[3] = newNaughtyDisk(posixDisk, nil, errDiskNotFound)
	expectedErr := fmt.Errorf("Unable to initialize format %s and %s", errSomeDiskOffline, errSomeDiskUnformatted)
	if err = healFormatXL(xl.storageDisks); err != nil {
		if err.Error() != expectedErr.Error() {
			t.Fatal("Got an unexpected error: ", err)
		}
	}
	removeRoots(fsDirs)

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// One disk is not found, heal corrupted disks should return nil
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	if err = obj.MakeBucket(getRandomBucketName()); err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= 2; i++ {
		if err = xl.storageDisks[i].DeleteFile(".minio.sys", "format.json"); err != nil {
			t.Fatal(err)
		}
	}
	if err = healFormatXL(xl.storageDisks); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}
	removeRoots(fsDirs)
}

// Tests undoes and validates if the undoing completes successfully.
func TestUndoMakeBucket(t *testing.T) {
	root, err := newTestConfig("us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	defer removeAll(root)

	nDisks := 16
	fsDirs, err := getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}
	defer removeRoots(fsDirs)

	endpoints, err := parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Remove format.json on 16 disks.
	obj, _, err := initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}

	bucketName := getRandomBucketName()
	if err = obj.MakeBucket(bucketName); err != nil {
		t.Fatal(err)
	}
	xl := obj.(*xlObjects)
	undoMakeBucket(xl.storageDisks, bucketName)

	// Validate if bucket was deleted properly.
	_, err = obj.GetBucketInfo(bucketName)
	if err != nil {
		err = errorCause(err)
		switch err.(type) {
		case BucketNotFound:
		default:
			t.Fatal(err)
		}
	}
}

// Tests quick healing of bucket and bucket metadata.
func TestQuickHeal(t *testing.T) {
	root, err := newTestConfig("us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	defer removeAll(root)

	nDisks := 16
	fsDirs, err := getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}
	defer removeRoots(fsDirs)

	endpoints, err := parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// Remove format.json on 16 disks.
	obj, _, err := initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}

	bucketName := getRandomBucketName()
	if err = obj.MakeBucket(bucketName); err != nil {
		t.Fatal(err)
	}

	xl := obj.(*xlObjects)
	for i := 0; i <= 2; i++ {
		if err = xl.storageDisks[i].DeleteVol(bucketName); err != nil {
			t.Fatal(err)
		}
	}

	// Heal the missing buckets.
	if err = quickHeal(xl.storageDisks, xl.writeQuorum); err != nil {
		t.Fatal(err)
	}

	// Validate if buckets were indeed healed.
	for i := 0; i <= 2; i++ {
		if _, err = xl.storageDisks[i].StatVol(bucketName); err != nil {
			t.Fatal(err)
		}
	}

	// Corrupt one of the disks to return unformatted disk.
	posixDisk, ok := xl.storageDisks[0].(*posix)
	if !ok {
		t.Fatal("storage disk is not *posix type")
	}
	xl.storageDisks[0] = newNaughtyDisk(posixDisk, nil, errUnformattedDisk)
	if err = quickHeal(xl.storageDisks, xl.writeQuorum); err != errUnformattedDisk {
		t.Fatal(err)
	}

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}
	defer removeRoots(fsDirs)

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// One disk is not found, heal corrupted disks should return nil
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	xl.storageDisks[0] = nil
	if err = quickHeal(xl.storageDisks, xl.writeQuorum); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}

	fsDirs, err = getRandomDisks(nDisks)
	if err != nil {
		t.Fatal(err)
	}
	defer removeRoots(fsDirs)

	endpoints, err = parseStorageEndpoints(fsDirs)
	if err != nil {
		t.Fatal(err)
	}

	// One disk is not found, heal corrupted disks should return nil
	obj, _, err = initObjectLayer(endpoints)
	if err != nil {
		t.Fatal(err)
	}
	xl = obj.(*xlObjects)
	// Corrupt one of the disks to return unformatted disk.
	posixDisk, ok = xl.storageDisks[0].(*posix)
	if !ok {
		t.Fatal("storage disk is not *posix type")
	}
	xl.storageDisks[0] = newNaughtyDisk(posixDisk, nil, errDiskNotFound)
	if err = quickHeal(xl.storageDisks, xl.writeQuorum); err != nil {
		t.Fatal("Got an unexpected error: ", err)
	}
}
