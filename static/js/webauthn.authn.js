// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

'use strict';

$('#login').submit(function(event) {
    event.preventDefault();

    if (this.checkValidity() === false) {
        this.classList.add('was-validated');
        return
    }

    const optionsRequest = {
        "username": this.username.value,
        "userVerification": this.userverification.value,
    }

    getAssertionOptions(optionsRequest)
        .then((options) => {
            return navigator.credentials.get({"publicKey": options})
        })
        .then((credential) => {
            return sendAssertionResult(credential)
        }).then(() => {
            window.location.href = "/"
        })
        .catch((error) => alert(error))
})

async function getAssertionOptions(optionsRequest) {
    const response = await fetch('/assertion/options', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(optionsRequest)
    });
    if (response.headers.get('Content-Type') !== 'application/json') {
        throw new TypeError("/assertion/options response header has unexpected Content-Type: " + response.headers.get('Content-Type'));
    }
    let optionsResponse = await response.json();
    if (optionsResponse.status !== 'ok') {
        throw new Error(`${optionsResponse.errorMessage}`);
    }
    optionsResponse.challenge = base64url.decode(optionsResponse.challenge);
    if (typeof optionsResponse.allowCredentials !== "undefined") {
        for (let i = 0; i < optionsResponse.allowCredentials.length; i++) {
            optionsResponse.allowCredentials[i].id = base64url.decode(optionsResponse.allowCredentials[i].id);
        }
    }
    return optionsResponse;
}

async function sendAssertionResult(credential) {
    let credentialResponse = {}
    credentialResponse['authenticatorData'] = base64url.encode(credential.response.authenticatorData)
    credentialResponse['clientDataJSON'] = base64url.encode(credential.response.clientDataJSON)
    credentialResponse['signature'] = base64url.encode(credential.response.signature)
    credentialResponse['userHandle'] = credential.response.userHandle
    let resultRequest = {}
    resultRequest['id'] = credential.id
    resultRequest['rawId'] = base64url.encode(credential.rawId)
    resultRequest["response"] = credentialResponse
    resultRequest['type'] = credential.type

    const response = await fetch('/assertion/result', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(resultRequest)
    });
    if (response.headers.get('Content-Type') !== 'application/json') {
        throw new TypeError("/assertion/result response header has unexpected Content-Type: " + response.headers.get('Content-Type'));
    }
    const resultResponse = await response.json();
    if (resultResponse.status !== 'ok')
        throw new Error(`${resultResponse.errorMessage}`);
    return resultResponse;
}
