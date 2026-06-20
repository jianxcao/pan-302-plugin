use prost::Message;

pub fn encode_response<T: Message>(value: &T) -> i64 {
    let mut bytes = Vec::new();
    if value.encode(&mut bytes).is_err() {
        return 0;
    }
    let len = bytes.len();
    let mut boxed = bytes.into_boxed_slice();
    let ptr = boxed.as_mut_ptr() as u32;
    std::mem::forget(boxed);
    ((ptr as i64) << 32) | len as i64
}

pub unsafe fn read_proto<T: Message + Default>(
    ptr: i32,
    len: i32,
) -> Result<T, String> {
    if ptr < 0 || len < 0 {
        return Err("invalid pointer".to_string());
    }
    let bytes = std::slice::from_raw_parts(ptr as *const u8, len as usize);
    T::decode(bytes).map_err(|error| error.to_string())
}

#[no_mangle]
pub extern "C" fn pan302_alloc(size: i32) -> i32 {
    if size <= 0 {
        return 0;
    }
    let mut buffer = vec![0_u8; size as usize].into_boxed_slice();
    let ptr = buffer.as_mut_ptr();
    std::mem::forget(buffer);
    ptr as i32
}

#[no_mangle]
pub unsafe extern "C" fn pan302_free(ptr: i32, len: i32) {
    if ptr <= 0 || len <= 0 {
        return;
    }
    let slice = std::ptr::slice_from_raw_parts_mut(ptr as *mut u8, len as usize);
    drop(Box::from_raw(slice));
}
