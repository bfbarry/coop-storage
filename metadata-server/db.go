package metadataserver
import (

)


type MetaObject struct {
	id string
	owner string
	fileType string
	fileName string
	deleteFlag bool

	// TODO: implement these for file integrity checks and multipart upload
	// offset int32
	// length int32
}

func (o *MetaObject) Create() (error) {
	//TODO
	return nil
}
