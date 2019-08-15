// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by Apache License 2.0 found in the LICENSE file.

'use strict';

class base64url {
    static encode(arraybuffer) {
        let s = String.fromCharCode.apply(null, new Uint8Array(arraybuffer))
        return window.btoa(s).replace(/\+/g, '-').replace(/\//g, '_');        
    }

    static decode(base64string) {
        let s = window.atob(base64string.replace(/-/g, '+').replace(/_/g, '/'))
        let bytes = Uint8Array.from(s, c=>c.charCodeAt(0))
        return bytes.buffer
    }
}