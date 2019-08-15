// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

'use strict';

$('#register').submit(function(event) {
    event.preventDefault();

    if (this.checkValidity() === false) {
        this.classList.add('was-validated');
        return
    }

    const optionsRequest = {
        "username": this.username.value,
        "displayName": this.displayName.value,
        "authenticatorSelection": {
            "authenticatorAttachment": this.authenticator.value,
            "requireResidentKey": (this.residentkey.value === "required"),
            "residentKey": this.residentkey.value,
            "userVerification": this.userverification.value,
        },
        "attestation": this.attestation.value,
    }

    getAttestationOptions(optionsRequest)
        .then((options) => {
            return navigator.credentials.create({"publicKey": options})
        })
        .then((credential) => {
            return sendAttestationResult(credential)
        })
        .then(() => {
            window.location.href = "/"
        })
        .catch((error) => alert(error))        
})

async function getAttestationOptions(optionsRequest) {
    const response = await fetch('/attestation/options', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(optionsRequest),
    });
    if (response.headers.get('Content-Type') !== 'application/json') {
        throw new TypeError("/attestation/options response header has unexpected Content-Type: " + response.headers.get('Content-Type'));
    }
    const optionsResponse = await response.json();
    if (optionsResponse.status !== 'ok')
        throw new Error(`${optionsResponse.errorMessage}`);
    optionsResponse.challenge = base64url.decode(optionsResponse.challenge);
    optionsResponse.user.id = base64url.decode(optionsResponse.user.id)
    if (typeof optionsResponse.excludeCredentials !== "undefined") {
        for (let i = 0; i < optionsResponse.excludeCredentials.length; i++) {
            optionsResponse.excludeCredentials[i].id = base64url.decode(optionsResponse.excludeCredentials[i].id);
        }
    }        
    return optionsResponse;
}

async function sendAttestationResult(credential) {
    let credentialResponse = {}
    credentialResponse['clientDataJSON'] = base64url.encode(credential.response.clientDataJSON)
    credentialResponse['attestationObject'] = base64url.encode(credential.response.attestationObject)
    let resultRequest = {}
    resultRequest['id'] = credential.id
    resultRequest['rawId'] = base64url.encode(credential.rawId)
    resultRequest["response"] = credentialResponse
    resultRequest['type'] = credential.type

    const response = await fetch('/attestation/result', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(resultRequest)
    });
    if (response.headers.get('Content-Type') !== 'application/json') {
        throw new TypeError("/attestation/result response header has unexpected Content-Type: " + response.headers.get('Content-Type'));
    }
    const resultResponse = await response.json();
    if (resultResponse.status !== 'ok')
        throw new Error(`${resultResponse.errorMessage}`);
    return resultResponse;
}
