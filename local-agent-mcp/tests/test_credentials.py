from bbs_go_mcp.credential_store import CredentialStore


def test_encrypted_file_store_round_trip_without_plaintext(tmp_path):
    store = CredentialStore(tmp_path, use_keyring=False)
    credential = store.save("secret-cookie")
    assert credential.cookie == "secret-cookie"
    raw = store.credentials_path.read_bytes()
    assert b"secret-cookie" not in raw
    assert b"secret-cookie" not in store.master_key_path.read_bytes()
    loaded = store.load()
    assert loaded is not None
    assert loaded.cookie == "secret-cookie"
    if store.credentials_path.stat().st_mode & 0o777 != 0o600:
        raise AssertionError("credential file permissions are too broad")
    store.delete()
    assert store.load() is None
