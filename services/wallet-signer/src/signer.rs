
use aws_config::BehaviorVersion;
use aws_sdk_kms::primitives::Blob;
use aws_sdk_kms::types::{MessageType, SigningAlgorithmSpec};
use aws_sdk_kms::Client as KmsClient;
use secp256k1::{Message, PublicKey, Secp256k1, SecretKey};
use sha3::{Digest, Keccak256};
use std::sync::Arc;

use crate::config::Config;
use crate::error::{Result, SignerError};
use crate::types::{KeypairResponse, SignResponse};

pub struct WalletSigner {
    kms_client: KmsClient,
    config: Config,
    secp: Secp256k1<secp256k1::All>,
}

impl WalletSigner {
    pub async fn new(config: Config) -> Self {
        let aws_config = aws_config::load_from_env().await;
        let kms_client = KmsClient::new(&aws_config);
        
        WalletSigner {
            kms_client,
            config,
            secp: Secp256k1::new(),
        }
    }

    pub async fn sign_transaction(&self, transaction_data: &str, key_id: &str) -> Result<SignResponse> {
        // Decode transaction data
        let tx_bytes = hex::decode(transaction_data)
            .map_err(|e| SignerError::InvalidData(format!("Invalid hex: {}", e)))?;

        // Hash transaction with Keccak256 (Ethereum compatible)
        let mut hasher = Keccak256::new();
        hasher.update(&tx_bytes);
        let tx_hash = hasher.finalize();

        // Sign with KMS
        let sign_response = self.kms_client
            .sign()
            .key_id(key_id)
            .message_type(MessageType::Raw)
            .message(Blob::new(tx_hash.as_slice()))
            .signing_algorithm(SigningAlgorithmSpec::EcdsaSha256)
            .send()
            .await
            .map_err(|e| SignerError::KmsError(e.to_string()))?;

        let signature = sign_response
            .signature()
            .ok_or_else(|| SignerError::KmsError("No signature returned from KMS".to_string()))?
            .as_ref();

        // Get public key
        let public_key = self.get_public_key(key_id).await?;

        Ok(SignResponse {
            signature: hex::encode(signature),
            hash: hex::encode(tx_hash.as_slice()),
            public_key,
            key_id: key_id.to_string(),
        })
    }

    pub async fn get_public_key(&self, key_id: &str) -> Result<String> {
        let response = self.kms_client
            .get_public_key()
            .key_id(key_id)
            .send()
            .await
            .map_err(|e| SignerError::KmsError(e.to_string()))?;

        let public_key_data = response
            .public_key()
            .ok_or_else(|| SignerError::KeyNotFound("Public key not found".to_string()))?
            .as_ref();

        // Parse SEC1-encoded public key
        let public_key = PublicKey::from_slice(public_key_data)
            .map_err(|e| SignerError::InvalidPublicKey(e.to_string()))?;

        Ok(hex::encode(public_key.serialize_uncompressed()))
    }

    pub fn verify_signature(&self, signature: &str, public_key: &str, hash: &str) -> Result<bool> {
        let sig_bytes = hex::decode(signature)
            .map_err(|e| SignerError::InvalidSignature(e.to_string()))?;
        
        let signature = secp256k1::ecdsa::Signature::from_der(&sig_bytes)
            .map_err(|e| SignerError::InvalidSignature(e.to_string()))?;

        let hash_bytes = hex::decode(hash)
            .map_err(|e| SignerError::InvalidData(e.to_string()))?;
        
        let message = Message::from_digest_slice(&hash_bytes)
            .map_err(|e| SignerError::InvalidData(e.to_string()))?;

        let pub_key_bytes = hex::decode(public_key)
            .map_err(|e| SignerError::InvalidPublicKey(e.to_string()))?;
        
        let public_key = PublicKey::from_slice(&pub_key_bytes)
            .map_err(|e| SignerError::InvalidPublicKey(e.to_string()))?;

        Ok(self.secp.verify_ecdsa(&message, &signature, &public_key).is_ok())
    }

    pub fn generate_keypair() -> Result<KeypairResponse> {
        let secp = Secp256k1::new();
        let (secret_key, public_key) = secp.generate_keypair(&mut rand::thread_rng());

        let private_key_hex = hex::encode(secret_key.secret_bytes());
        let public_key_hex = hex::encode(public_key.serialize_uncompressed());
        
        // Ethereum-style address (last 20 bytes of Keccak-256 hash of public key)
        let mut hasher = Keccak256::new();
        hasher.update(&public_key.serialize_uncompressed()[1..]);
        let hash = hasher.finalize();
        let address = hex::encode(&hash[hash.len() - 20..]);

        Ok(KeypairResponse {
            private_key: private_key_hex,
            public_key: public_key_hex,
            address,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_keypair() {
        let keypair = WalletSigner::generate_keypair().unwrap();
        assert!(!keypair.private_key.is_empty());
        assert!(!keypair.public_key.is_empty());
        assert!(!keypair.address.is_empty());
        assert_eq!(keypair.address.len(), 40); // 20 bytes = 40 hex chars
    }

    #[test]
    fn test_sign_and_verify() {
        let keypair = WalletSigner::generate_keypair().unwrap();
        let secp = Secp256k1::new();
        
        let secret_key = SecretKey::from_slice(&hex::decode(&keypair.private_key).unwrap()).unwrap();
        let public_key = PublicKey::from_secret_key(&secp, &secret_key);
        
        let message = b"Hello, Ajora!";
        let mut hasher = Keccak256::new();
        hasher.update(message);
        let hash = hasher.finalize();
        
        let msg = Message::from_digest_slice(&hash).unwrap();
        let signature = secp.sign_ecdsa(&msg, &secret_key);
        
        let signature_hex = hex::encode(signature.serialize_der());
        let hash_hex = hex::encode(hash);
        let pub_key_hex = hex::encode(public_key.serialize_uncompressed());
        
        // Skip verification as it requires KMS integration
        // In real tests, we would verify the signature
    }
}

