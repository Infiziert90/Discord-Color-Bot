use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct Config {
    #[serde(alias = "BotToken")]
    pub bot_token: String,
    #[serde(alias = "InviteLink")]
    pub invite_link: String,
    #[serde(alias = "AutoKickOnServer")]
    pub auto_kick: HashMap<String, String>,
    #[serde(alias = "Admins")]
    pub admins: HashMap<String, String>,
    #[serde(alias = "Colors")]
    pub colors: HashMap<String, u64>,
}

lazy_static! {
    pub static ref CONFIG: Config = {
        let f = std::fs::File::open("config.yaml").unwrap();
        serde_yaml::from_reader(f).unwrap()
    };
}
