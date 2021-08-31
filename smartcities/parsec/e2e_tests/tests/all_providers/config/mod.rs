// Copyright 2020 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0
use e2e_tests::TestClient;
use log::{error, info};
use parsec_client::core::interface::operations::list_providers::Uuid;
use parsec_client::core::interface::operations::psa_algorithm::Hash;
use parsec_client::core::interface::operations::psa_algorithm::{Algorithm, AsymmetricSignature};
use parsec_client::core::interface::operations::psa_key_attributes::{
    Attributes, Lifetime, Policy, Type, UsageFlags,
};
use parsec_client::core::interface::requests::ResponseStatus;
use std::env;
use std::fs;
use std::path::PathBuf;
use std::process::Command;
use std::thread;
use std::time::Duration;

const CONFIG_TOMLS_FOLDER: &str = "tests/all_providers/config/tomls";
const SERVICE_CONFIG_PATH: &str = "provider_cfg/tmp_config.toml";

fn set_config(filename: &str) {
    info!("Changing service configuration file to {}", filename);
    let config_path = PathBuf::from(SERVICE_CONFIG_PATH);
    let mut new_config = env::current_dir() // this is the root of the crate for tests
        .unwrap();
    new_config.push(CONFIG_TOMLS_FOLDER);
    new_config.push(filename);
    if !new_config.exists() {
        error!("Configuration file {} does not exist", filename);
        panic!();
    }

    let _ = fs::copy(new_config, config_path).unwrap();
}

fn reload_service() {
    info!("Reloading Parsec service");

    let _ = Command::new("pkill")
        .arg("-SIGHUP")
        .arg("parsec")
        .output()
        .expect("Reloading service failed");

    // wait for the service to restart
    thread::sleep(Duration::from_secs(2));
}

#[test]
fn list_providers() {
    set_config("list_providers_1.toml");
    reload_service();

    let mut client = TestClient::new();
    let providers = client.list_providers().unwrap();
    let uuids: Vec<Uuid> = providers.iter().map(|p| p.uuid).collect();
    assert_eq!(
        uuids,
        vec![
            Uuid::parse_str("1c1139dc-ad7c-47dc-ad6b-db6fdb466552").unwrap(), // Mbed crypto provider
            Uuid::parse_str("1e4954a4-ff21-46d3-ab0c-661eeb667e1d").unwrap(), // Tpm provider
            Uuid::parse_str("30e39502-eba6-4d60-a4af-c518b7f5e38f").unwrap(), // Pkcs11 provider
            Uuid::parse_str("b8ba81e2-e9f7-4bdd-b096-a29d0019960c").unwrap(), // CryptoAuthLib provider
            Uuid::parse_str("47049873-2a43-4845-9d72-831eab668784").unwrap(), // Core provider
        ]
    );

    set_config("list_providers_2.toml");
    reload_service();

    let providers = client.list_providers().unwrap();
    let uuids: Vec<Uuid> = providers.iter().map(|p| p.uuid).collect();
    assert_eq!(
        uuids,
        vec![
            Uuid::parse_str("30e39502-eba6-4d60-a4af-c518b7f5e38f").unwrap(), // Pkcs11 provider
            Uuid::parse_str("1c1139dc-ad7c-47dc-ad6b-db6fdb466552").unwrap(), // Mbed crypto provider
            Uuid::parse_str("1e4954a4-ff21-46d3-ab0c-661eeb667e1d").unwrap(), // Tpm provider
            Uuid::parse_str("b8ba81e2-e9f7-4bdd-b096-a29d0019960c").unwrap(), // CryptoAuthLib provider
            Uuid::parse_str("47049873-2a43-4845-9d72-831eab668784").unwrap(), // Core provider
        ]
    );
}

#[cfg(feature = "pkcs11-provider")]
#[test]
fn pkcs11_verify_software() {
    use sha2::{Digest, Sha256};
    set_config("pkcs11_software.toml");
    reload_service();

    let mut client = TestClient::new();
    let key_name = String::from("pkcs11_verify_software");

    let mut hasher = Sha256::new();
    hasher.update(b"Bob wrote this message.");
    let hash = hasher.finalize().to_vec();

    client.generate_rsa_sign_key(key_name.clone()).unwrap();

    let signature = client
        .sign_with_rsa_sha256(key_name.clone(), hash.clone())
        .unwrap();
    client
        .verify_with_rsa_sha256(key_name, hash, signature)
        .unwrap();
}

#[cfg(feature = "pkcs11-provider")]
#[test]
fn pkcs11_verify_software_ecc() {
    use sha2::{Digest, Sha256};
    set_config("pkcs11_software.toml");
    reload_service();

    let mut client = TestClient::new();
    let key_name = String::from("pkcs11_verify_software_ecc");

    let mut hasher = Sha256::new();
    hasher.update(b"Bob wrote this message.");
    let hash = hasher.finalize().to_vec();

    client
        .generate_ecc_key_pair_secpr1_ecdsa_sha256(key_name.clone())
        .unwrap();

    let signature = client
        .sign_with_ecdsa_sha256(key_name.clone(), hash.clone())
        .unwrap();
    client
        .verify_with_ecdsa_sha256(key_name, hash, signature)
        .unwrap();
}

#[cfg(feature = "pkcs11-provider")]
#[test]
fn pkcs11_encrypt_software() {
    set_config("pkcs11_software.toml");
    reload_service();

    let mut client = TestClient::new();
    let key_name = String::from("pkcs11_verify_software");
    let plaintext_msg = [
        0x69, 0x3E, 0xDB, 0x1B, 0x22, 0x79, 0x03, 0xF4, 0xC0, 0xBF, 0xD6, 0x91, 0x76, 0x37, 0x84,
        0xA2, 0x94, 0x8E, 0x92, 0x50, 0x35, 0xC2, 0x8C, 0x5C, 0x3C, 0xCA, 0xFE, 0x18, 0xE8, 0x81,
        0x37, 0x78,
    ];
    client
        .generate_rsa_encryption_keys_rsaoaep_sha1(key_name.clone())
        .unwrap();
    let ciphertext = client
        .asymmetric_encrypt_message_with_rsaoaep_sha1(
            key_name.clone(),
            plaintext_msg.to_vec(),
            vec![],
        )
        .unwrap();
    let plaintext = client
        .asymmetric_decrypt_message_with_rsaoaep_sha1(key_name, ciphertext, vec![])
        .unwrap();
    assert_eq!(&plaintext_msg[..], &plaintext[..]);
}

#[test]
fn no_tpm_support() {
    set_config("no_tpm_support.toml");
    // The service should still start, without the TPM provider.
    reload_service();

    let mut client = TestClient::new();
    let providers = client.list_providers().unwrap();
    let uuids: Vec<Uuid> = providers.iter().map(|p| p.uuid).collect();
    assert_eq!(
        uuids,
        vec![
            Uuid::parse_str("1c1139dc-ad7c-47dc-ad6b-db6fdb466552").unwrap(), // Mbed crypto provider
            Uuid::parse_str("30e39502-eba6-4d60-a4af-c518b7f5e38f").unwrap(), // Pkcs11 provider
            Uuid::parse_str("47049873-2a43-4845-9d72-831eab668784").unwrap(), // Core provider
        ]
    );
}

#[test]
fn various_fields() {
    set_config("various_field_check.toml");
    reload_service();

    env::set_var("PARSEC_SERVICE_ENDPOINT", "unix:/tmp/toto.sock");

    let mut client = TestClient::new();
    // Try to send a bit less than 1KiB, should work
    let _ = client
        .hash_compute(Hash::Sha256, &vec![0xDD; 1019])
        .unwrap();
    // Try to send 1KiB and one byte, should fail
    assert_eq!(
        client
            .hash_compute(Hash::Sha256, &vec![0xDD; 1025])
            .unwrap_err(),
        ResponseStatus::BodySizeExceedsLimit
    );

    let _ = client.generate_bytes(1024).unwrap();
    assert_eq!(
        client.generate_bytes(1025).unwrap_err(),
        ResponseStatus::ResponseTooLarge
    );

    env::set_var("PARSEC_SERVICE_ENDPOINT", "unix:/tmp/parsec.sock");
}

#[test]
fn allow_export() {
    set_config("allow_export.toml");
    reload_service();

    let mut client = TestClient::new();
    assert_eq!(
        client
            .generate_key(
                "allow_export".to_string(),
                Attributes {
                    lifetime: Lifetime::Persistent,
                    key_type: Type::RsaKeyPair,
                    bits: 1024,
                    policy: Policy {
                        usage_flags: UsageFlags {
                            sign_hash: true,
                            verify_hash: true,
                            sign_message: true,
                            verify_message: true,
                            // Should not be allowed by configuration
                            export: true,
                            encrypt: false,
                            decrypt: false,
                            cache: false,
                            copy: false,
                            derive: false,
                        },
                        permitted_algorithms: Algorithm::AsymmetricSignature(
                            AsymmetricSignature::RsaPkcs1v15Sign {
                                hash_alg: Hash::Sha256.into(),
                            },
                        ),
                    },
                },
            )
            .unwrap_err(),
        ResponseStatus::PsaErrorNotPermitted
    );
}

#[test]
fn ts_pkcs11_cross() {
    use super::cross::{import_and_verify, import_and_verify_ecc, setup_sign, setup_sign_ecc};
    use parsec_client::core::interface::requests::ProviderId;
    set_config("ts_pkcs11_cross.toml");
    reload_service();

    let key_name = String::from("ts_pkcs11_sign_cross");
    let (mut client, pub_key, signature) = setup_sign(ProviderId::TrustedService, key_name.clone());
    import_and_verify(
        &mut client,
        ProviderId::Pkcs11,
        key_name.clone(),
        pub_key.clone(),
        signature.clone(),
    );

    let key_name_ecc = String::from("ts_pkcs11_sign_cross_ecc");
    let (mut client, pub_key, signature) =
        setup_sign_ecc(ProviderId::TrustedService, key_name_ecc.clone());
    import_and_verify_ecc(
        &mut client,
        ProviderId::Pkcs11,
        key_name_ecc.clone(),
        pub_key.clone(),
        signature.clone(),
    );

    let key_name = String::from("pkcs11_ts_sign_cross");
    let (mut client, pub_key, signature) = setup_sign(ProviderId::Pkcs11, key_name.clone());
    import_and_verify(
        &mut client,
        ProviderId::TrustedService,
        key_name.clone(),
        pub_key.clone(),
        signature.clone(),
    );

    let key_name_ecc = String::from("pkcs11_ts_sign_cross_ecc");
    let (mut client, pub_key, signature) = setup_sign_ecc(ProviderId::Pkcs11, key_name_ecc.clone());
    import_and_verify_ecc(
        &mut client,
        ProviderId::TrustedService,
        key_name_ecc.clone(),
        pub_key.clone(),
        signature.clone(),
    );
}

#[test]
fn no_user_pin() {
    set_config("no_user_pin.toml");
    // The service should still start, without the user pin.
    reload_service();

    let mut client = TestClient::new();
    let _ = client.ping().unwrap();
}

#[test]
fn no_slot_number() {
    set_config("no_slot_number.toml");
    // The service should still start, without the slot number.
    reload_service();

    let mut client = TestClient::new();
    let _ = client.ping().unwrap();
}
