package main
import (
	"fmt"
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
)


// TODO: don;t save ID in column, waste of space
type MetaObject struct {
	// KEY
	ID         string `json:"id"`
	//VALUES
	Owner      string `json:"owner"`
	FileType   string `json:"fileType"`
	FileName   string `json:"fileName"`
	DeleteFlag bool   `json:"deleteFlag"`

	// TODO: implement these for file integrity checks and multipart upload
	// offset int32
	// length int32
}

func (o *MetaObject) Read() (error) {
	oKey := []byte(fmt.Sprintf("objid:%s", o.ID))
	fmt.Printf("MetaObject.Read: ID=%s, searching for key: %s\n", o.ID, string(oKey))
	var currMeta MetaObject
	currMetaJSON, err := DBInst.Read(oKey)
	if err != nil {
		fmt.Printf("MetaObject.Read: Failed to read key %s: %v\n", string(oKey), err)
		return err
	}

	errj := json.Unmarshal([]byte(currMetaJSON), &currMeta) 

	if errj != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	o.Owner  = currMeta.Owner
	o.FileType = currMeta.FileType
	o.FileName = currMeta.FileName
	o.DeleteFlag = currMeta.DeleteFlag
	
	return nil
}

func (o *MetaObject) Write() (error) {
	// object index
	jsonStr, err := json.Marshal(o)
	if err != nil {
		fmt.Printf("MetaObject.Write: Failed to marshal object: %v\n", err)
		return err
	}
	oKey := fmt.Sprintf("objid:%s", o.ID)
	fmt.Printf("MetaObject.Write: ID=%s, writing key: %s\n", o.ID, oKey)
	if err := DBInst.Update([]byte(oKey), jsonStr); err != nil {
		fmt.Printf("MetaObject.Write: Failed to update key %s: %v\n", oKey, err)
		return err
	}

	// user index
	uKey := []byte(fmt.Sprintf("user:%s", o.Owner))
	var currFiles []string
	currFilesJSON, err := DBInst.Read(uKey)
	if err == badger.ErrKeyNotFound {
		currFilesJSON = []byte("[]")
	}

	errj := json.Unmarshal([]byte(currFilesJSON), &currFiles) 

	if errj != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	currFiles = append(currFiles, o.ID)
	currFilesJSON, err = json.Marshal(currFiles)
	DBInst.Update(uKey, currFilesJSON)

	return nil
}