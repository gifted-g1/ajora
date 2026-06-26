
use std::env;

#[derive(Debug, Clone)]
pub struct Config {
    pub kms_key_id: String,
    pub aws_region: String,
    pub aws_access_key_id: Option<String>,
    pub aws_secret_access_key: Option<String>,
    pub signing_algorithm: String,
}

impl Config {
    pub fn from_env() -> Result<Self, Box<dyn std::error::Error>> {
        dotenvy::dotenv().ok();

        Ok(Config {
            kms_key_id: env::var("KMS_KEY_ID")
                .unwrap_or_else(|_| "alias/ajora-signing-key".to_string()),
            aws_region: env::var("AWS_REGION")
                .unwrap_or_else(|_| "us-west-2".to_string()),
            aws_access_key_id: env::var("AWS_ACCESS_KEY_ID").ok(),
            aws_secret_access_key: env::var("AWS_SECRET_ACCESS_KEY").ok(),
            signing_algorithm: env::var("SIGNING_ALGORITHM")
                .unwrap_or_else(|_| "ECDSA_SHA_256".to_string()),
        })
    }
}

