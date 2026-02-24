use std::fs;

fn read_config() -> String {
    let content = fs::read_to_string("config.toml").unwrap();
    let value = content.parse::<i32>().unwrap();
    format!("value: {}", value)
}

fn main() {
    println!("{}", read_config());
}
