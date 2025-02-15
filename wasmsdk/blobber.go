//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/core/sys"
	"github.com/hack-pad/safejs"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const FileOperationInsert = "insert"

func listObjects(allocationID string, remotePath string, offset, pageLimit int) (*sdk.ListResult, error) {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	return alloc.ListDir(remotePath, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
}

func listObjectsFromAuthTicket(allocationID, authTicket, lookupHash string, offset, pageLimit int) (*sdk.ListResult, error) {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return nil, err
	}
	return alloc.ListDirFromAuthTicket(authTicket, lookupHash, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
}

func cancelUpload(allocationID, remotePath string) error {
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return err
	}
	return allocationObj.CancelUpload(remotePath)
}

func pauseUpload(allocationID, remotePath string) error {
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return err
	}
	return allocationObj.PauseUpload(remotePath)
}

func createDir(allocationID, remotePath string) error {
	if len(allocationID) == 0 {
		return RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return RequiredArg("remotePath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return err
	}

	return allocationObj.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationCreateDir,
			RemotePath:    remotePath,
		},
	})
}

// getFileStats get file stats from blobbers
func getFileStats(allocationID, remotePath string) ([]*sdk.FileStats, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	fileStats, err := allocationObj.GetFileStats(remotePath)
	if err != nil {
		return nil, err
	}

	var stats []*sdk.FileStats

	for _, it := range fileStats {
		stats = append(stats, it)
	}

	return stats, nil
}

// updateBlobberSettings expects settings JSON of type sdk.Blobber
func updateBlobberSettings(blobberSettingsJson string) (*transaction.Transaction, error) {
	var blobberSettings sdk.Blobber
	err := json.Unmarshal([]byte(blobberSettingsJson), &blobberSettings)
	if err != nil {
		sdkLogger.Error(err)
		return nil, err
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS,
		InputArgs: blobberSettings,
	}

	_, _, _, txn, err := sdk.StorageSmartContractTxn(sn)
	return txn, err
}

// Delete delete file from blobbers
func Delete(allocationID, remotePath string) (*FileCommandResponse, error) {

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationDelete,
			RemotePath:    remotePath,
		},
	})
	sdkLogger.Info(remotePath + " deleted")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Rename rename a file existing already on dStorage. Only the allocation's owner can rename a file.
func Rename(allocationID, remotePath, destName string) (*FileCommandResponse, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	if len(destName) == 0 {
		return nil, RequiredArg("destName")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationRename,
			RemotePath:    remotePath,
			DestName:      destName,
		},
	})

	if err != nil {
		PrintError(err.Error())
		return nil, err
	}
	sdkLogger.Info(remotePath + " renamed")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Copy copy file to another folder path on blobbers
func Copy(allocationID, remotePath, destPath string) (*FileCommandResponse, error) {

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	if len(destPath) == 0 {
		return nil, RequiredArg("destPath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationCopy,
			RemotePath:    remotePath,
			DestPath:      destPath,
		},
	})

	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	sdkLogger.Info(remotePath + " copied")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Move move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.
func Move(allocationID, remotePath, destPath string) (*FileCommandResponse, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	if len(destPath) == 0 {
		return nil, RequiredArg("destPath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationMove,
			RemotePath:    remotePath,
			DestPath:      destPath,
		},
	})

	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	sdkLogger.Info(remotePath + " moved")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Share  generate an authtoken that provides authorization to the holder to the specified file on the remotepath.
func Share(allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool, availableAfter string) (string, error) {

	if len(allocationID) == 0 {
		return "", RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return "", RequiredArg("remotePath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return "", err
	}

	refType := fileref.DIRECTORY

	sdkLogger.Info("getting filestats")
	statsMap, err := allocationObj.GetFileStats(remotePath)
	if err != nil {
		PrintError("Error in getting information about the object." + err.Error())
		return "", err
	}

	for _, v := range statsMap {
		if v != nil {
			refType = fileref.FILE
			break
		}
	}

	var fileName string
	_, fileName = pathutil.Split(remotePath)

	if revoke {
		err := allocationObj.RevokeShare(remotePath, clientID)
		if err != nil {
			PrintError(err.Error())
			return "", err
		}
		sdkLogger.Info("Share revoked for client " + clientID)
		return "", nil
	}

	availableAt := time.Now()

	if len(availableAfter) > 0 {
		aa, err := common.ParseTime(availableAt, availableAfter)
		if err != nil {
			PrintError(err.Error())
			return "", err
		}
		availableAt = *aa
	}

	ref, err := allocationObj.GetAuthTicket(remotePath, fileName, refType, clientID, encryptionPublicKey, int64(expiration), &availableAt)
	if err != nil {
		PrintError(err.Error())
		return "", err
	}
	sdkLogger.Info("Auth token :" + ref)

	return ref, nil

}

// MultiOperation - do copy, move, delete and createdir operation together
// ## Inputs
//   - allocationID
//   - jsonMultiDownloadOptions: Json Array of MultiDownloadOption.
//	 - authTicket
//  - callbackFuncName: callback function name Invoke with totalBytes, completedBytes, objURL, err
// ## Outputs
//   - json string of array of DownloadCommandResponse
// 	 - error

func multiDownload(allocationID, jsonMultiDownloadOptions, authTicket, callbackFuncName string) (string, error) {
	sdkLogger.Info("starting multidownload")
	wg := &sync.WaitGroup{}
	useCallback := false
	if callbackFuncName != "" {
		useCallback = true
	}
	var options []*MultiDownloadOption
	err := json.Unmarshal([]byte(jsonMultiDownloadOptions), &options)
	if err != nil {
		return "", err
	}
	var alloc *sdk.Allocation
	if authTicket == "" {
		alloc, err = getAllocation(allocationID)
	} else {
		alloc, err = sdk.GetAllocationFromAuthTicket(authTicket)
	}
	if err != nil {
		return "", err
	}
	allStatusBar := make([]*StatusBar, len(options))
	wg.Add(len(options))
	for ind, option := range options {
		fileName := strings.Replace(path.Base(option.RemotePath), "/", "-", -1)
		localPath := allocationID + "_" + fileName
		option.LocalPath = localPath
		statusBar := &StatusBar{wg: wg}
		allStatusBar[ind] = statusBar
		if useCallback {
			callback := js.Global().Get(callbackFuncName)
			statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
				callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
			}
		}
		var mf sys.File
		if option.DownloadToDisk {
			terminateWorkersWithAllocation(alloc)
			mf, err = jsbridge.NewFileWriter(fileName)
			if err != nil {
				PrintError(err.Error())
				return "", err
			}
		} else {
			statusBar.localPath = localPath
			fs, _ := sys.Files.Open(localPath)
			mf, _ = fs.(*sys.MemFile)
		}

		var downloader sdk.Downloader
		if option.DownloadOp == 1 {
			downloader, err = sdk.CreateDownloader(allocationID, localPath, option.RemotePath,
				sdk.WithAllocation(alloc),
				sdk.WithAuthticket(authTicket, option.RemoteLookupHash),
				sdk.WithOnlyThumbnail(false),
				sdk.WithBlocks(0, 0, option.NumBlocks),
				sdk.WithFileHandler(mf),
			)
		} else {
			downloader, err = sdk.CreateDownloader(allocationID, localPath, option.RemotePath,
				sdk.WithAllocation(alloc),
				sdk.WithAuthticket(authTicket, option.RemoteLookupHash),
				sdk.WithOnlyThumbnail(true),
				sdk.WithBlocks(0, 0, option.NumBlocks),
				sdk.WithFileHandler(mf),
			)
		}
		if err != nil {
			PrintError(err.Error())
			return "", err
		}
		defer sys.Files.Remove(option.LocalPath) //nolint
		downloader.Start(statusBar, ind == len(options)-1)
	}
	wg.Wait()
	resp := make([]DownloadCommandResponse, len(options))

	for ind, statusBar := range allStatusBar {
		statusResponse := DownloadCommandResponse{}
		if !statusBar.success {
			statusResponse.CommandSuccess = false
			statusResponse.Error = "Download failed: " + statusBar.err.Error()
		} else {
			statusResponse.CommandSuccess = true
			statusResponse.FileName = options[ind].RemoteFileName
			statusResponse.Url = statusBar.objURL
		}
		resp[ind] = statusResponse
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(respBytes), nil
}

type BulkUploadOption struct {
	AllocationID string `json:"allocationId,omitempty"`
	RemotePath   string `json:"remotePath,omitempty"`

	ThumbnailBytes jsbridge.Bytes `json:"thumbnailBytes,omitempty"`
	Encrypt        bool           `json:"encrypt,omitempty"`
	IsWebstreaming bool           `json:"webstreaming,omitempty"`
	IsUpdate       bool           `json:"isUpdate,omitempty"`
	IsRepair       bool           `json:"isRepair,omitempty"`

	NumBlocks         int    `json:"numBlocks,omitempty"`
	FileSize          int64  `json:"fileSize,omitempty"`
	ReadChunkFuncName string `json:"readChunkFuncName,omitempty"`
	CallbackFuncName  string `json:"callbackFuncName,omitempty"`
	Md5HashFuncName   string `json:"md5HashFuncName,omitempty"`
	MimeType          string `json:"mimeType,omitempty"`
	MemoryStorer      bool   `json:"memoryStorer,omitempty"`
}

type BulkUploadResult struct {
	RemotePath string `json:"remotePath,omitempty"`
	Success    bool   `json:"success,omitempty"`
	Error      string `json:"error,omitempty"`
}
type MultiUploadResult struct {
	Success bool   `json:"success,omitempty"`
	Error   string `json:"error,omitempty"`
}

type MultiOperationOption struct {
	OperationType string `json:"operationType,omitempty"`
	RemotePath    string `json:"remotePath,omitempty"`
	DestName      string `json:"destName,omitempty"` // Required only for rename operation
	DestPath      string `json:"destPath,omitempty"` // Required for copy and move operation`
}

type MultiDownloadOption struct {
	RemotePath       string `json:"remotePath"`
	LocalPath        string `json:"localPath,omitempty"`
	DownloadOp       int    `json:"downloadOp"`
	NumBlocks        int    `json:"numBlocks"`
	RemoteFileName   string `json:"remoteFileName"`             //Required only for file download with auth ticket
	RemoteLookupHash string `json:"remoteLookupHash,omitempty"` //Required only for file download with auth ticket
	DownloadToDisk   bool   `json:"downloadToDisk"`
}

// MultiOperation - do copy, move, delete and createdir operation together
// ## Inputs
//   - allocationID
//   - jsonMultiUploadOptions: Json Array of MultiOperationOption. eg: "[{"operationType":"move","remotePath":"/README.md","destPath":"/folder1/"},{"operationType":"delete","remotePath":"/t3.txt"}]"
//
// ## Outputs
//   - error
func MultiOperation(allocationID string, jsonMultiUploadOptions string) error {
	if allocationID == "" {
		return errors.New("AllocationID is required")
	}

	if jsonMultiUploadOptions == "" {
		return errors.New("operations are empty")
	}

	var options []MultiOperationOption
	err := json.Unmarshal([]byte(jsonMultiUploadOptions), &options)
	if err != nil {
		sdkLogger.Info("error unmarshalling")
		return err
	}
	totalOp := len(options)
	operations := make([]sdk.OperationRequest, totalOp)
	for idx, op := range options {
		operations[idx] = sdk.OperationRequest{
			OperationType: op.OperationType,
			RemotePath:    op.RemotePath,
			DestName:      op.DestName,
			DestPath:      op.DestPath,
		}
	}
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return allocationObj.DoMultiOperation(operations)
}

func bulkUpload(jsonBulkUploadOptions string) ([]BulkUploadResult, error) {
	var options []BulkUploadOption
	err := json.Unmarshal([]byte(jsonBulkUploadOptions), &options)
	if err != nil {
		return nil, err
	}
	n := len(options)
	wait := make(chan BulkUploadResult, 1)

	for _, option := range options {
		go func(o BulkUploadOption) {
			result := BulkUploadResult{
				RemotePath: o.RemotePath,
			}
			defer func() { wait <- result }()

			ok, err := uploadWithJsFuncs(o.AllocationID, o.RemotePath,
				o.ReadChunkFuncName,
				o.FileSize,
				o.ThumbnailBytes.Buffer,
				o.IsWebstreaming,
				o.Encrypt,
				o.IsUpdate,
				o.IsRepair,
				o.NumBlocks,
				o.CallbackFuncName)
			result.Success = ok
			if err != nil {
				result.Error = err.Error()
				result.Success = false
			}

		}(option)

	}

	results := make([]BulkUploadResult, 0, n)
	for i := 0; i < n; i++ {
		result := <-wait
		results = append(results, result)
	}

	return results, nil
}

// set upload mode, default is medium, for low set 0, for high set 2
func setUploadMode(mode int) {
	switch mode {
	case 0:
		sdk.SetUploadMode(sdk.UploadModeLow)
	case 1:
		sdk.SetUploadMode(sdk.UploadModeMedium)
	case 2:
		sdk.SetUploadMode(sdk.UploadModeHigh)
	}
}

func multiUpload(jsonBulkUploadOptions string) (MultiUploadResult, error) {
	var options []BulkUploadOption
	result := MultiUploadResult{}
	err := json.Unmarshal([]byte(jsonBulkUploadOptions), &options)
	if err != nil {
		result.Error = "Error in unmarshaling json"
		result.Success = false
		return result, err
	}
	n := len(options)
	if n == 0 {
		result.Error = "No files to upload"
		result.Success = false
		return result, errors.New("There are nothing to upload")
	}
	allocationID := options[0].AllocationID
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		result.Error = "Error fetching the allocation"
		result.Success = false
		return result, errors.New("Error fetching the allocation")
	}
	err = addWebWorkers(allocationObj)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, err
	}

	operationRequests := make([]sdk.OperationRequest, n)
	for idx, option := range options {
		wg := &sync.WaitGroup{}
		statusBar := &StatusBar{wg: wg}
		callbackFuncName := option.CallbackFuncName
		if callbackFuncName != "" {
			callback := js.Global().Get(callbackFuncName)
			statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
				callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
			}
		}
		wg.Add(1)
		encrypt := option.Encrypt
		remotePath := option.RemotePath
		fileReader, err := jsbridge.NewFileReader(option.ReadChunkFuncName, option.FileSize, allocationObj.GetChunkReadSize(encrypt))
		if err != nil {
			result.Error = "Error in file operation"
			result.Success = false
			return result, err
		}
		mimeType := option.MimeType
		localPath := remotePath
		remotePath = zboxutil.RemoteClean(remotePath)
		isabs := zboxutil.IsRemoteAbs(remotePath)
		if !isabs {
			err = errors.New("invalid_path: Path should be valid and absolute")
			result.Error = err.Error()
			result.Success = false
			return result, err
		}
		fullRemotePath := zboxutil.GetFullRemotePath(localPath, remotePath)

		_, fileName := pathutil.Split(fullRemotePath)

		if mimeType == "" {
			mimeType, err = zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
			if err != nil {
				result.Error = "Error in file operation"
				result.Success = false
				return result, err
			}
		}

		fileMeta := sdk.FileMeta{
			Path:       localPath,
			ActualSize: option.FileSize,
			MimeType:   mimeType,
			RemoteName: fileName,
			RemotePath: fullRemotePath,
		}
		numBlocks := option.NumBlocks
		if numBlocks <= 1 {
			numBlocks = 100
		}

		options := []sdk.ChunkedUploadOption{
			sdk.WithThumbnail(option.ThumbnailBytes.Buffer),
			sdk.WithEncrypt(encrypt),
			sdk.WithStatusCallback(statusBar),
			sdk.WithChunkNumber(numBlocks),
		}
		if option.MemoryStorer {
			options = append(options, sdk.WithProgressStorer(&chunkedUploadProgressStorer{
				list: make(map[string]*sdk.UploadProgress),
			}))
		}
		if option.Md5HashFuncName != "" {
			fileHasher := newFileHasher(option.Md5HashFuncName)
			options = append(options, sdk.WithFileHasher(fileHasher))
		}
		operationRequests[idx] = sdk.OperationRequest{
			FileMeta:       fileMeta,
			FileReader:     fileReader,
			OperationType:  FileOperationInsert,
			Opts:           options,
			Workdir:        "/",
			IsWebstreaming: option.IsWebstreaming,
		}

	}
	err = allocationObj.DoMultiOperation(operationRequests)
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result, err
	}
	result.Success = true
	return result, nil
}

func uploadWithJsFuncs(allocationID, remotePath string, readChunkFuncName string, fileSize int64, thumbnailBytes []byte, webStreaming, encrypt, isUpdate, isRepair bool, numBlocks int, callbackFuncName string) (bool, error) {

	if len(allocationID) == 0 {
		return false, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return false, RequiredArg("remotePath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return false, err
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	if callbackFuncName != "" {
		callback := js.Global().Get(callbackFuncName)
		statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
			callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
		}
	}
	wg.Add(1)

	fileReader, err := jsbridge.NewFileReader(readChunkFuncName, fileSize, allocationObj.GetChunkReadSize(encrypt))
	if err != nil {
		return false, err
	}

	localPath := remotePath

	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = errors.New("invalid_path: Path should be valid and absolute")
		return false, err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)

	_, fileName := pathutil.Split(remotePath)

	mimeType, err := zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
	if err != nil {
		return false, err
	}

	fileMeta := sdk.FileMeta{
		Path:       localPath,
		ActualSize: fileSize,
		MimeType:   mimeType,
		RemoteName: fileName,
		RemotePath: remotePath,
	}

	if numBlocks < 1 {
		numBlocks = 100
	}
	if allocationObj.DataShards > 7 {
		numBlocks = 50
	}

	ChunkedUpload, err := sdk.CreateChunkedUpload(context.TODO(), "/", allocationObj, fileMeta, fileReader, isUpdate, isRepair, webStreaming, zboxutil.NewConnectionId(),
		sdk.WithThumbnail(thumbnailBytes),
		sdk.WithEncrypt(encrypt),
		sdk.WithStatusCallback(statusBar),
		sdk.WithChunkNumber(numBlocks))
	if err != nil {
		return false, err
	}

	err = ChunkedUpload.Start()

	if err != nil {
		PrintError("Upload failed.", err)
		return false, err
	}

	wg.Wait()
	if !statusBar.success {
		return false, errors.New("upload failed: unknown")
	}

	return true, nil
}

// upload upload file
func upload(allocationID, remotePath string, fileBytes, thumbnailBytes []byte, webStreaming, encrypt, isUpdate, isRepair bool, numBlocks int) (*FileCommandResponse, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	fileReader := bytes.NewReader(fileBytes)

	localPath := remotePath

	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = errors.New("invalid_path: Path should be valid and absolute")
		return nil, err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)

	_, fileName := pathutil.Split(remotePath)

	mimeType, err := zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
	if err != nil {
		return nil, err
	}

	fileMeta := sdk.FileMeta{
		Path:       localPath,
		ActualSize: int64(len(fileBytes)),
		MimeType:   mimeType,
		RemoteName: fileName,
		RemotePath: remotePath,
	}

	if numBlocks < 1 {
		numBlocks = 100
	}

	ChunkedUpload, err := sdk.CreateChunkedUpload(context.TODO(), "/", allocationObj, fileMeta, fileReader, isUpdate, isRepair, webStreaming,
		zboxutil.NewConnectionId(),
		sdk.WithThumbnail(thumbnailBytes),
		sdk.WithEncrypt(encrypt),
		sdk.WithStatusCallback(statusBar),
		sdk.WithChunkNumber(numBlocks))
	if err != nil {
		return nil, err
	}

	err = ChunkedUpload.Start()

	if err != nil {
		PrintError("Upload failed.", err)
		return nil, err
	}
	wg.Wait()
	if !statusBar.success {
		return nil, errors.New("upload failed: unknown")
	}

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// download download file blocks
func downloadBlocks(allocId string, remotePath, authTicket, lookupHash string, startBlock, endBlock int64) ([]byte, error) {

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return nil, RequiredArg("remotePath/authTicket")
	}

	alloc, err := getAllocation(allocId)

	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	var (
		wg        = &sync.WaitGroup{}
		statusBar = &StatusBar{wg: wg}
	)

	fileName := strings.Replace(path.Base(remotePath), "/", "-", -1)
	localPath := alloc.ID + "-" + fmt.Sprintf("%v-%s", startBlock, fileName)

	fs, err := sys.Files.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("could not open local file: %v", err)
	}

	mf, _ := fs.(*sys.MemFile)
	if mf == nil {
		return nil, fmt.Errorf("invalid memfile")
	}

	defer sys.Files.Remove(localPath) //nolint

	wg.Add(1)
	if authTicket != "" {
		err = alloc.DownloadByBlocksToFileHandlerFromAuthTicket(mf, authTicket, lookupHash, startBlock, endBlock, 100, remotePath, false, statusBar, true)
	} else {
		err = alloc.DownloadByBlocksToFileHandler(
			mf,
			remotePath,
			startBlock,
			endBlock,
			100,
			false,
			statusBar, true)
	}
	if err != nil {
		return nil, err
	}
	wg.Wait()
	return mf.Buffer, nil
}

// GetBlobbersList get list of active blobbers, and format them as array json string
func getBlobbers(stakable bool) ([]*sdk.Blobber, error) {
	blobbs, err := sdk.GetBlobbers(true, stakable)
	if err != nil {
		return nil, err
	}
	return blobbs, err
}

func repairAllocation(allocationID string) error {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	statusBar := sdk.NewRepairBar(allocationID)
	if statusBar == nil {
		return errors.New("repair already in progress")
	}
	err = alloc.RepairAlloc(statusBar)
	if err != nil {
		return err
	}
	statusBar.Wait()
	return statusBar.CheckError()
}

func checkAllocStatus(allocationID string) (string, error) {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}
	status, blobberStatus, err := alloc.CheckAllocStatus()
	var statusStr string
	switch status {
	case sdk.Repair:
		statusStr = "repair"
	case sdk.Broken:
		statusStr = "broken"
	default:
		statusStr = "ok"
	}
	statusResult := CheckStatusResult{
		Status:        statusStr,
		Err:           err,
		BlobberStatus: blobberStatus,
	}
	statusBytes, err := json.Marshal(statusResult)
	if err != nil {
		return "", err
	}

	return string(statusBytes), err
}

func skipStatusCheck(allocationID string, checkStatus bool) error {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	alloc.SetCheckStatus(checkStatus)
	return nil
}

func terminateWorkers(allocationID string) {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return
	}
	for _, blobber := range alloc.Blobbers {
		jsbridge.RemoveWorker(blobber.ID)
	}
}

func terminateWorkersWithAllocation(alloc *sdk.Allocation) {
	for _, blobber := range alloc.Blobbers {
		jsbridge.RemoveWorker(blobber.ID)
	}
}

func createWorkers(allocationID string) error {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return addWebWorkers(alloc)
}

func startListener() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	selfWorker, err := jsbridge.NewSelfWorker()
	if err != nil {
		return err
	}
	safeVal, _ := safejs.ValueOf("startListener")
	selfWorker.PostMessage(safeVal, nil) //nolint:errcheck

	listener, err := selfWorker.Listen(ctx)
	if err != nil {
		return err
	}
	sdk.InitHasherMap()
	for event := range listener {
		data, err := event.Data()
		if err != nil {
			PrintError("Error in getting data from event", err)
			return err
		}
		sdk.ProcessEventData(data)
	}

	return nil
}
