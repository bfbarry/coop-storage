package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

//--------------------------------------------------------------------------------------------
//Define MetaDirectory and MetaObject Types
//--------------------------------------------------------------------------------------------

type MetaDirectory struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`     //just the folder name
	ParentID string `json:"parentId"` //"" if root directory
}

// TODO: don;t save ID in column, waste of space
type MetaObject struct {
	// KEY
	ID string `json:"id"`
	//VALUES
	Owner      string `json:"owner"`
	FileType   string `json:"fileType"`
	FileName   string `json:"fileName"`
	DirID      string `json:"dirId"` //points to MetaDirectory.ID
	DeleteFlag bool   `json:"deleteFlag"`
	Version    int    `json:"version"`
}

// --------------------------------------------------------------------------------------------
// CRUD for Meta Directory
// --------------------------------------------------------------------------------------------
func (d *MetaDirectory) Create() error {
	dKey := NewDBKey(Directory, d.ID)

	_, err := DBInst.Read(dKey)
	if err == nil {
		return fmt.Errorf("directory with ID %s already exists", d.ID)
	}
	if err != badger.ErrKeyNotFound {
		return err
	}

	jsonStr, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("MetaDirectory.Create: failed to marshal: %w", err)
	}

	return DBInst.Update(dKey, jsonStr)
}

func (d *MetaDirectory) Read() error {
	if d.ID == "" {
		return fmt.Errorf("MetaDirectory needs an id to be Read")
	}
	dKey := NewDBKey(Directory, d.ID)
	data, err := DBInst.Read(dKey)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, d)
}

func (d *MetaDirectory) Delete() error {
	// first check no children exist
	children, err := GetDirectoryChildren(d.Owner, d.ID)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return fmt.Errorf("directory %s is not empty", d.ID)
	}
	return DBInst.Delete(NewDBKey(Directory, d.ID))
}

//--------------------------------------------------------------------------------------------
//CRUD for Meta Object
//--------------------------------------------------------------------------------------------

func (o *MetaObject) Create() error {
	oKey := NewDBKey(Object, o.ID)

	// Check if object with this ID already exists
	_, err := DBInst.Read(oKey)
	if err == nil {
		log.Printf("MetaObject.Create: Object with ID %s already exists\n", o.ID)
		return fmt.Errorf("object with ID %s already exists", o.ID)
	}
	if err != badger.ErrKeyNotFound {
		log.Printf("MetaObject.Create: Failed to check for existing key %s: %v\n", oKey, err)
		return err
	}

	// object index
	jsonStr, err := json.Marshal(o)
	if err != nil {
		log.Printf("MetaObject.Create: Failed to marshal object: %v\n", err)
		return err
	}
	log.Printf("MetaObject.Create: ID=%s, writing key: %s\n", o.ID, oKey)

	if err := UpdateUserIndex(o.Owner, o.FileName, o.DirID, o.ID, "", "", Add); err != nil {
		return err
	}

	// TODO: use newer UpdateObject instead
	if err := DBInst.Update(oKey, jsonStr); err != nil {
		log.Printf("MetaObject.Create: Failed to update key %s: %v\n", oKey, err)
		return err
	}

	return nil
}

func (o *MetaObject) Read() error {
	if o.ID == "" {
		return fmt.Errorf("MetaObject needs an id to be Read")
	}
	oKey := NewDBKey(Object, o.ID)
	// TODO: this seems very wrong, try to deserialize directly to self instead, given
	var currMeta MetaObject
	currMetaJSON, err := DBInst.Read(oKey)
	if err != nil {
		log.Printf("MetaObject.Read: Failed to read key %s: %v\n", string(oKey), err)
		return err
	}

	if err := json.Unmarshal([]byte(currMetaJSON), &currMeta); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	o.Owner = currMeta.Owner
	o.FileType = currMeta.FileType
	o.FileName = currMeta.FileName
	o.DeleteFlag = currMeta.DeleteFlag
	o.DirID = currMeta.DirID
	o.Version = currMeta.Version

	return nil
}

func (o *MetaObject) Update() error {
	// TODO: check if here we also need concurrency protection
	if o.DeleteFlag {
		return o.Delete()
	}

	oKey := NewDBKey(Object, o.ID)
	var currMeta MetaObject
	currMetaJSON, err := DBInst.Read(oKey)
	if err != nil {
		log.Printf("MetaObject.Update: Failed to read key %s: %v\n", string(oKey), err)
		return err
	}

	if err := json.Unmarshal([]byte(currMetaJSON), &currMeta); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ErrOnWrite
	}

	//NOTE: done first to check for uniqueness
	if err := UpdateUserIndex(o.Owner, o.FileName, o.DirID, o.ID, currMeta.FileName, currMeta.DirID, Modify); err != nil {
		return err
	}

	if err := DBInst.UpdateObj(o); err != nil {
		log.Printf("failed to update %v", err)
		return err
	}

	return nil
}

func (o *MetaObject) Delete() error {
	if err := UpdateUserIndex(o.Owner, o.FileName, o.DirID, "", "", "", Remove); err != nil {
		return err
	}
	if err := DBInst.Delete(NewDBKey(Object, o.ID)); err != nil {
		UpdateUserIndex(o.Owner, o.FileName, o.DirID, o.ID, "", "", Add)
		return err
	}
	return nil
}

//--------------------------------------------------------------------------------------------
//UTILS
//--------------------------------------------------------------------------------------------

var (
	userIndexMu sync.Map
)

type UpdateArrayMode int

const (
	Add UpdateArrayMode = iota
	Modify
	Remove
)

func StartGarbageCollector() error {
	DBInst.db.RunValueLogGC(0.5)
	return nil
}

func (d *MetaDirectory) Rename(newName string) error {
	d.Name = newName
	dKey := NewDBKey(Directory, d.ID)
	jsonStr, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("MetaDirectory.Rename: failed to marshal: %w", err)
	}
	return DBInst.Update(dKey, jsonStr)
}

func GetDirectoryChildren(owner string, dirID string) ([]string, error) {
	uKey := NewDBKey(User, owner)
	objectMap := make(map[string]string)

	objectMapJSON, err := DBInst.Read(uKey)
	if err != nil && err != badger.ErrKeyNotFound {
		return nil, err
	} else if err == badger.ErrKeyNotFound {
		return []string{}, nil
	}

	if err := json.Unmarshal([]byte(objectMapJSON), &objectMap); err != nil {
		return nil, err
	}

	prefix := dirID + "/"
	children := []string{}
	for k := range objectMap {
		if strings.HasPrefix(k, prefix) {
			children = append(children, k)
		}
	}
	return children, nil
}

func getUserMutex(user string) *sync.Mutex {
	mu, _ := userIndexMu.LoadOrStore(user, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// always call this first to ensure fName uniqueness
func UpdateUserIndex(user string, fName string, dirID string, objId string, oldFname string, oldDirID string, mode UpdateArrayMode) error {
	mu := getUserMutex(user)
	mu.Lock()
	defer mu.Unlock()

	uKey := NewDBKey(User, user)
	objectMap := make(map[string]string)

	objectMapJSON, err := DBInst.Read(uKey)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == badger.ErrKeyNotFound {
		objectMapJSON = []byte("{}")
	}

	if err := json.Unmarshal([]byte(objectMapJSON), &objectMap); err != nil {
		return fmt.Errorf("UpdateUserIndex: failed to unmarshal: %w", err)
	}

	// key is now "dirID/fileName" instead of just "fileName"
	fullKey := dirID + "/" + fName
	oldKey := oldDirID + "/" + oldFname

	switch mode {
	case Add:
		if _, ok := objectMap[fullKey]; ok {
			return fmt.Errorf("UpdateUserIndex: file already exists in this directory")
		}
		objectMap[fullKey] = objId
	case Modify:
		if fullKey != oldKey {
			if _, ok := objectMap[fullKey]; ok {
				return fmt.Errorf("UpdateUserIndex: file already exists in this directory")
			}
		}
		delete(objectMap, oldKey)
		objectMap[fullKey] = objId
	case Remove:
		delete(objectMap, fullKey)
	}

	objectMapJSON, err = json.Marshal(objectMap)
	if err != nil {
		return err
	}
	return DBInst.Update(uKey, objectMapJSON)
}
