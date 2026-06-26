
mod config;
mod error;
mod signer;
mod types;

use actix_web::{web, App, HttpServer, HttpResponse, middleware};
use tracing::{info, error};
use tracing_subscriber;

use crate::config::Config;
use crate::signer::WalletSigner;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt::init();

    // Load configuration
    let config = Config::from_env().expect("Failed to load configuration");
    info!("Starting Wallet Signer service");

    // Initialize signer
    let signer = web::Data::new(WalletSigner::new(config).await);

    HttpServer::new(move || {
        App::new()
            .app_data(signer.clone())
            .wrap(middleware::Logger::default())
            .wrap(middleware::Compress::default())
            .service(
                web::scope("/api/v1/signer")
                    .route("/sign", web::post().to(sign_transaction))
                    .route("/verify", web::post().to(verify_signature))
                    .route("/generate-keypair", web::post().to(generate_keypair))
                    .route("/health", web::get().to(health_check))
            )
            .route("/health", web::get().to(health_check))
    })
    .bind(("0.0.0.0", 8088))?
    .run()
    .await
}

async fn sign_transaction(
    signer: web::Data<WalletSigner>,
    req: web::Json<types::SignRequest>,
) -> HttpResponse {
    match signer.sign_transaction(&req.transaction_data, &req.key_id).await {
        Ok(response) => HttpResponse::Ok().json(response),
        Err(e) => {
            error!("Signing failed: {}", e);
            HttpResponse::InternalServerError().json(serde_json::json!({
                "error": format!("Failed to sign transaction: {}", e)
            }))
        }
    }
}

async fn verify_signature(
    signer: web::Data<WalletSigner>,
    req: web::Json<types::VerifyRequest>,
) -> HttpResponse {
    match signer.verify_signature(&req.signature, &req.public_key, &req.hash) {
        Ok(valid) => HttpResponse::Ok().json(serde_json::json!({
            "valid": valid,
            "message": if valid { "Signature verified successfully" } else { "Invalid signature" }
        })),
        Err(e) => {
            error!("Verification failed: {}", e);
            HttpResponse::InternalServerError().json(serde_json::json!({
                "error": format!("Failed to verify signature: {}", e)
            }))
        }
    }
}

async fn generate_keypair() -> HttpResponse {
    match WalletSigner::generate_keypair() {
        Ok(keypair) => HttpResponse::Ok().json(keypair),
        Err(e) => {
            error!("Keypair generation failed: {}", e);
            HttpResponse::InternalServerError().json(serde_json::json!({
                "error": format!("Failed to generate keypair: {}", e)
            }))
        }
    }
}

async fn health_check() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "healthy",
        "service": "wallet-signer",
        "version": "1.0.0",
        "timestamp": chrono::Utc::now().to_rfc3339()
    }))
}

