package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	libvirt "github.com/libvirt/libvirt-go"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	// parse arg
	numb := flag.Int("numb", 700, "numer of goroutines")
	libvirtURI := flag.String("uri", "qemu:///system", "libvirt URI connection")
	flag.Parse()
	// this is for goroutines sync
	bench := *numb
	err_messages := make(chan error)
	var wg sync.WaitGroup
	// create libvirt connection depending on var
	fmt.Println("[DEBUG:] creating libvirt connection...")
	virConn, err := libvirt.NewConnect(*libvirtURI)
	if err != nil {
		panic(err)
	}

	fmt.Println("[DEBUG:] created libvirt connection")
	defer virConn.Close()

	// here is the benchmark
	wg.Add(bench)
	for i := 0; i < bench; i++ {
		go func() {
			defer wg.Done()
			err_messages <- benchmarkLibvirt(virConn)
		}()
	}
	go func() {
		for i := range err_messages {
			fmt.Println(i)
		}
	}()

	wg.Wait()

}

func benchmarkLibvirt(virConn *libvirt.Connect) error {
	fmt.Println("creating libvirt volume...")
	volume, err := createVolume(virConn)
	if err != nil {
		fmt.Println("Error by craeting vol %s", err)
		return (err)
	}
	fmt.Println("creating libvirt volume created !!!")
	// copy some content to Volume
	err = writeToVolume(virConn, &volume)
	if err != nil {
		fmt.Printf("Error by wrting to volume :%s", err)
		return (err)
	}
	defer func() {
		volume.Free()
	}()
	// download vol content to tmpFile
	err = copyVolumeToTmpFile(virConn, volume)
	if err != nil {
		fmt.Printf("Error by copying from vol to file %s", err)
		return (err)
	}
	return nil
}

func writeToVolume(virConn *libvirt.Connect, volume *libvirt.StorageVol) error {
	var bytesCopied int64

	reader := strings.NewReader("soffffme io.Reader stream to be read ffffff aas\n")
	info, err := volume.GetInfo()

	if err != nil {
		return fmt.Errorf("Error retrieving info for volume: %s", err)
	}

	size := info.Capacity
	stream, err := virConn.NewStream(0)
	if err != nil {
		return err
	}

	defer func() {
		stream.Free()
	}()
	sio := NewStreamIO(*stream)
	if err := volume.Upload(stream, 0, size, 0); err != nil {
		stream.Abort()
		return fmt.Errorf("Error while uploading volume %s", err)
	}

	bytesCopied, err = io.Copy(sio, reader)

	if err != nil {
		return fmt.Errorf("Error while copying source to volume %s", err)
	}
	log.Printf("%d bytes uploaded\n", bytesCopied)
	return nil
}

func testStorageVolXML(volName, poolPath string) string {
	defName := volName
	if defName == "" {
		defName = time.Now().String()
	}
	fmt.Printf("volname: %s", defName)
	return `<volume>
        <name>` + defName + `</name>
        <allocation>0</allocation>
        <capacity unit="M">10</capacity>
        <target>
          <path>` + "/" + poolPath + "/" + defName + `</path>
          <permissions>
            <owner>107</owner>
            <group>107</group>
            <mode>0744</mode>
            <label>testLabel0</label>
          </permissions>
        </target>
      </volume>`

}

func createVolume(virConn *libvirt.Connect) (libvirt.StorageVol, error) {
	poolPath := "/var/lib/libvirt/images"
	var volEmpty libvirt.StorageVol

	pool, err := virConn.LookupStoragePoolByName("default")
	if err != nil {
		return volEmpty, fmt.Errorf("can't find storage pool")
	}
	defer pool.Free()

	vol, err := pool.StorageVolCreateXML(testStorageVolXML(randString(10), poolPath), 0)
	if err != nil {
		panic(err)
		return volEmpty, err
	}

	return *vol, nil
}

func copyVolumeToTmpFile(virConn *libvirt.Connect, volume libvirt.StorageVol) error {
	var bytesCopied int64
	info, err := volume.GetInfo()
	if err != nil {
		return fmt.Errorf("Error retrieving info for volume: %s", err)
	}

	// create tmp file for the ISO
	tmpFile, err := ioutil.TempFile("/tmp/performance", "cloudinit")
	if err != nil {
		return fmt.Errorf("Cannot create tmp file: %s", err)
	}

	// download ISO file
	stream, err := virConn.NewStream(0)
	if err != nil {
		return fmt.Errorf("Stream Creation error: %s", err)
	}

	defer func() {
		stream.Free()
	}()

	err = volume.Download(stream, 0, info.Capacity, 0)
	if err != nil {
		stream.Abort()
		return fmt.Errorf("-> Volume Download Error: %s", err)
	}

	sio := NewStreamIO(*stream)

	bytesCopied, err = io.Copy(tmpFile, sio)
	if err != nil {
		return fmt.Errorf("Error while copying remote volume to local disk: %s", err)
	}
	fmt.Printf("Bytes to tmpfile copied %d", bytesCopied)
	return nil
}

func randString(n int) string {

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
