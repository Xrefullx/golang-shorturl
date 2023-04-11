syntax="proto3";

option go_package = "internal/grpc/proto";

package grpc;

message GetRequest{
string short_id = 1;
}
message GetResponse{
string src_url = 1;
string error = 2;
}

message GetListRequest{
string user_id =1;
}

message GetListItem{
string short_url = 1;
string src_url = 2;
}

message GetListResponse{
repeated GetListItem list = 1;
string error=2;
}

message SaveRequest{
string src_url = 1;
string user_id = 2;
}

message SaveResponse{
string short_url = 1;
string error = 2;
}

message SaveListItem{
string correlation_id = 1;
string url = 2;
}

message SaveListRequest{
repeated SaveListItem list = 1; // src urls
string user_id = 2;
}
message SaveListResponse{
repeated SaveListItem list = 1; // shortened urls
string error = 2;
}

message DelListRequest{
repeated string list =1;
string user_id = 2;
}
message DelListResponse{
string error = 1;
}

service URLs{
rpc Get(GetRequest) returns (GetResponse);
rpc GetList(GetListRequest) returns (GetListResponse);
rpc Save(SaveRequest) returns (SaveResponse);
rpc SaveList(SaveListRequest) returns (SaveListResponse);
rpc DelList(DelListRequest) returns (DelListResponse);
}