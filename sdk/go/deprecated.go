package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// Deprecated: Use Driver.List instead.
func DriverList() (*pb.DriverListResponse, error) {
	return Driver.List()
}

// Deprecated: Use Driver.Read instead.
func DriverRead(req *pb.DriverReadRequest) (*pb.DriverReadResponse, error) {
	return Driver.Read(req)
}

// Deprecated: Use Driver.Link instead.
func DriverLink(req *pb.DriverLinkRequest) (*pb.DriverLinkResponse, error) {
	return Driver.Link(req)
}

// Deprecated: Use Driver.Mkdir instead.
func DriverMkdir(req *pb.DriverMkdirRequest) (*pb.DriverObject, error) {
	return Driver.Mkdir(req)
}

// Deprecated: Use Driver.Rename instead.
func DriverRename(req *pb.DriverRenameRequest) (*pb.DriverRenameResponse, error) {
	return Driver.Rename(req)
}

// Deprecated: Use Driver.Delete instead.
func DriverDelete(req *pb.DriverDeleteRequest) (*pb.DriverDeleteResponse, error) {
	return Driver.Delete(req)
}

// Deprecated: Use Driver.Move instead.
func DriverMove(req *pb.DriverMoveCopyRequest) (*pb.DriverMoveCopyResponse, error) {
	return Driver.Move(req)
}

// Deprecated: Use Driver.Copy instead.
func DriverCopy(req *pb.DriverMoveCopyRequest) (*pb.DriverMoveCopyResponse, error) {
	return Driver.Copy(req)
}

// Deprecated: Use Strm.Write instead.
func StrmWrite(req *pb.StrmWriteRequest) (*pb.StrmOperationResult, error) {
	return Strm.Write(req)
}

// Deprecated: Use Strm.Delete instead.
func StrmDelete(req *pb.StrmDeleteRequest) (*pb.StrmOperationResult, error) {
	return Strm.Delete(req)
}

// Deprecated: Use Config.Read instead.
func ConfigRead() (*pb.ConfigReadResponse, error) {
	return Config.Read()
}

// Deprecated: Use Config.Write instead.
func ConfigWrite(req *pb.ConfigWriteRequest) (*pb.ConfigWriteResponse, error) {
	return Config.Write(req)
}

// Deprecated: Use Logger.Write instead.
func Log(level, message string, fields map[string]string) {
	Logger.Write(level, message, fields)
}

// Deprecated: Use HTTP.Request instead.
func RequestHTTP(request *pb.HTTPRequestArgs) (*pb.HTTPResponseData, error) {
	return HTTP.Request(request)
}

// Deprecated: Use Driver.HasCapability instead.
func HasCapability(driver *pb.DriverInfo, cap string) bool {
	return Driver.HasCapability(driver, cap)
}
