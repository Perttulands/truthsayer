pub fn validate(input: &str) -> bool {
    if input.is_empty() {
        panic!("input must not be empty");
    }
    true
}

pub fn process(data: &[u8]) {
    if data.len() > 1024 {
        panic!("data too large");
    }
}
