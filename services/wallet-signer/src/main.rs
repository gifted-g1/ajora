
use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use serde::{Deserialize, Serialize};
use sha3::{Keccak256, Digest};
use secp256k1::{Secp256k1, SecretKey, PublicKey, Message};
use std::sync::Arc;
use aws_sdk_kms::Client as KmsClient;
use aws_config::{load_from_env, BehaviorVersion};
use tracing::{info, error, instrument};
use tracing_subscriber;

#[derive(Debug, Serialize, Deserialize)]
struct SignRequest {
    transaction_data: String,
    key_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SignResponse {
    signature: String,
    hash: String,
    public_key: String,
}

struct AppState {
    kms_client: KmsClient,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    tracing_subscriber::fmt::init();
    
    let aws_config = load_from_env(BehaviorVersion::latest()).await;
    let kms_client = KmsClient::new(&aws_config);
    
    let app_state = web::Data::new(AppState {
        kms_client,
    });

    info!("Starting Wallet Signer service on port 8080");

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .wrap(middleware::Logger::default())
            .wrap(middleware::Compress::default())
            .service(
                web::scope("/api/v1/signer")
                    .route("/sign", web::post().to(sign_transaction))
                    .route("/verify", web::post().to(verify_signature))
                    .route("/generate-keypair", web::post().to(generate_keypair))
            )
            .route("/health", web::get().to(health_check))
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}

#[instrument(skip(state))]
async fn sign_transaction(
    state: web::Data<AppState>,
    request: web::Json<SignRequest>,
) -> HttpResponse {
    let tx_data = match hex::decode(&request.transaction_data) {
        Ok(data) => data,
        Err(e) => {
            error!("Failed to decode transaction data: {}", e);
            return HttpResponse::BadRequest().json(serde_json::json!({
                "error": "Invalid transaction data format"
            }));
        }
    };

    let mut hasher = Keccak256::new();
    hasher.update(&tx_data);
    let tx_hash = hasher.finalize();

    let sign_response = match state.kms_client
        .sign()
        .key_id(&request.key_id)
        .message(aws_sdk_kms::types::MessageType::Raw)
        .message(tx_hash.as_slice())
        .signing_algorithm(aws_sdk_kms::types::SigningAlgorithmSpec::EcdsaSha256)
        .send()
        .await
    {
        Ok(response) => response,
        Err(e) => {
            error!("KMS signing failed: {}", e);
            return HttpResponse::InternalServerError().json(serde_json::json!({
                "error": "Failed to sign transaction"
            }));
        }
    };

    let signature = match sign_response.signature {
        Some(sig) => hex::encode(sig.as_ref()),
        None => {
            return HttpResponse::InternalServerError().json(serde_json::json!({
                "error": "No signature returned from KMS"
            }));
        }
    };

    let public_key = match get_public_key(&state.kms_client, &request.key_id).await {
        Ok(key) => key,
        Err(e) => {
            error!("Failed to get public key: {}", e);
            return HttpResponse::InternalServerError().json(serde_json::json!({
                "error": "Failed to get public key"
            }));
        }
    };

    HttpResponse::Ok().json(SignResponse {
        signature,
        hash: hex::encode(tx_hash.as_slice()),
        public_key,
    })
}

async fn get_public_key(kms_client: &KmsClient, key_id: &str) -> Result<String, Box<dyn std::error::Error>> {
    let response = kms_client
        .get_public_key()
        .key_id(key_id)
        .send()
        .await?;

    let public_key_data = match response.public_key {
        Some(pub_key) => pub_key,
        None => return Err("No public key returned".into()),
    };

    let public_key = secp256k1::PublicKey::from_slice(public_key_data.as_ref())?;
    Ok(hex::encode(public_key.serialize_uncompressed()))
}

async fn verify_signature(
    state: web::Data<AppState>,
    request: web::Json<VerifyRequest>,
) -> HttpResponse {
    let sig = match secp256k1::ecdsa::Signature::from_str(&request.signature) {
        Ok(sig) => sig,
        Err(e) => {
            return HttpResponse::BadRequest().json(serde_json::json!({
                "error": format!("Invalid signature: {}", e)
            }));
        }
    };

    let public_key = match secp256k1::PublicKey::from_str(&request.public_key) {
        Ok(key) => key,
        Err(e) => {
            return HttpResponse::BadRequest().json(serde_json::json!({
                "error": format!("Invalid public key: {}", e)
            }));
        }
    };

    let message = Message::from_digest_slice(&hex::decode(&request.hash).unwrap()).unwrap();
    let secp = Secp256k1::new();

    let is_valid = secp.verify_ecdsa(&message, &sig, &public_key).is_ok();

    HttpResponse::Ok().json(serde_json::json!({
        "valid": is_valid,
        "message": if is_valid { "Signature verified successfully" } else { "Invalid signature" }
    }))
}

async fn generate_keypair() -> HttpResponse {
    let secp = Secp256k1::new();
    let (secret_key, public_key) = secp.generate_keypair(&mut rand::thread_rng());

    HttpResponse::Ok().json(serde_json::json!({
        "private_key": hex::encode(secret_key.secret_bytes()),
        "public_key": hex::encode(public_key.serialize_uncompressed()),
        "address": hex::encode(public_key.serialize_uncompressed()[1..]),
    }))
}

async fn health_check() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "healthy",
        "service": "wallet-signer",
        "version": "1.0.0"
    }))
}

#[derive(Debug, Deserialize)]
struct VerifyRequest {
    signature: String,
    public_key: String,
    hash: String,
}

