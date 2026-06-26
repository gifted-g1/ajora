
use thiserror::Error;

#[derive(Error, Debug)]
pub enum SignerError {
    #[error("KMS error: {0}")]
    KmsError(String),
    
    #[error("Invalid transaction data: {0}")]
    InvalidData(String),
    
    #[error("Invalid signature: {0}")]
    InvalidSignature(String),
    
    #[error("Invalid public key: {0}")]
    InvalidPublicKey(String),
    
    #[error("Key not found: {0}")]
    KeyNotFound(String),
    
    #[error("AWS error: {0}")]
    AwsError(#[from] aws_sdk_kms::Error),
    
    #[error("Serialization error: {0}")]
    SerializationError(#[from] serde_json::Error),
    
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}

pub type Result<T> = std::result::Result<T, SignerError>;

