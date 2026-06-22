fn main() {
    let proto_dir = "../../proto";
    let proto_files = &[
        "../../proto/plugin/v1/abi.proto",
        "../../proto/plugin/v1/driver.proto",
        "../../proto/plugin/v1/strm.proto",
        "../../proto/plugin/v1/events.proto",
        "../../proto/plugin/v1/config.proto",
        "../../proto/plugin/v1/log.proto",
        "../../proto/plugin/v1/notify.proto",
        "../../proto/plugin/v1/http.proto",
        "../../proto/plugin/v1/media.proto",
        "../../proto/plugin/v1/lifecycle.proto",
        "../../proto/plugin/v1/route.proto",
    ];

    for proto_file in proto_files {
        println!("cargo:rerun-if-changed={}", proto_file);
    }

    prost_build::compile_protos(proto_files, &[proto_dir]).unwrap();
}
