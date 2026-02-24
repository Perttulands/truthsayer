fn dangerous() {
    let ptr = 0x1234 as *const i32;
    unsafe {
        println!("{}", *ptr);
    }
}

fn also_dangerous() {
    unsafe {
        std::ptr::null::<i32>().read();
    }
}
