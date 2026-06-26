
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct SignRequest {
    pub transaction_data: String,
    pub key_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SignResponse {
    pub signature: String,
    pub hash: String,
    pub public_key: String,
    pub key_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct VerifyRequest {
    pub signature: String,
    pub public_key: String,
    pub hash: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct KeypairResponse {
    pub private_key: String,
    pub public_key: String,
    pub address: String,
}

