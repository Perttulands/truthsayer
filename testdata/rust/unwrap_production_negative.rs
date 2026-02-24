use std::fs;

fn read_config() -> Result<String, Box<dyn std::error::Error>> {
    let content = fs::read_to_string("config.toml")
        .expect("failed to read config.toml");
    let value = content.parse::<i32>().unwrap_or(0);
    Ok(format!("value: {}", value))
}
