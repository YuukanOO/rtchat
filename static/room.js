'use strict';

(async function() {
    // Will contain every peers we're connected to
    const peers = {};

    let stream = null;

    // Try to retrieve a stream given audio / video constraints but allow to fail
    // so a user can join as a guest only.
    try {
        stream = await navigator.mediaDevices.getUserMedia({
            video: true,
            audio: true,
        });
    } catch {
        console.error('could not retrieve stream matching audio / video constraints');
    }

    // Update the local preview element source
    document.querySelector('video.local').srcObject = stream;

    // Let's open the websocket connection for this particular room
    const ws = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws/" + config.roomID, config.roomCred);

    // Upon close, show an alert and go back to the web root
    ws.onclose = function() {        
        for(const id in peers) {
            peers[id].close();
        }
        
        alert('communication closed, nothing to see anymore');
    }

    ws.onmessage = async function(e) {
        const msg = JSON.parse(e.data);

        console.info(msg);
    
        if(msg.joined) {
            // New user has joined, let's starts an RTCPeerConnection for this user
            // and make an offer.
            const peer = await createPeer(msg.joined.id);

            peer.oniceconnectionstatechange = function() {
                // TODO: handle connection state change and restart ice?
                // https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Session_lifetime#ICE_restart
                console.info(peer.iceConnectionState);
                // sendOffer(msg.joined.id, peer, { iceRestart: true });
            }

            sendOffer(msg.joined.id, peer);
        }
    
        if(msg.left) {
            removePeer(msg.left.id);
        }
    
        if(msg.offer) {
            // An offer has been made by someone else, create an RTCPeerConnection
            // and sends an answer
            const peer = peers[msg.from] || await createPeer(msg.from);
            await peer.setRemoteDescription(msg.offer);
            const answer = await peer.createAnswer();
            await peer.setLocalDescription(answer);
    
            ws.send(JSON.stringify({
                answer,
                to: msg.from,
            }));
        }

        if(msg.answer) {
            const peer = peers[msg.from];
            peer.setRemoteDescription(msg.answer);
        }

        if(msg.ice) {
            const peer = peers[msg.from];
            peer.addIceCandidate(msg.ice);
        }
    }

    /**
     * Sends an offer using the given peer. It will affect the local description
     * as well as signalying the offer on the websocket.
     */
    async function sendOffer(to, peer, offerOptions) {
        const offer = await peer.createOffer(offerOptions);
        await peer.setLocalDescription(offer);
        ws.send(JSON.stringify({
            offer,
            to,
        }));
    }

    function findVideoElement(id) {
        return document.querySelector('video[data-id="'+id+'"]');
    }

    /**
     * Create a peer connection and append it to the collection of peers.
     * It will also add local tracks, attach events and send an offer.
     */
    async function createPeer(id) {
        const peer = new RTCPeerConnection(config); // Config here comes from the html template
        
        let videoEle = findVideoElement(id);

        if(!videoEle) {
            videoEle = document.createElement('video');
            videoEle.classList.add('videos__peer');
            videoEle.dataset.id = id;
            videoEle.autoplay = true;
            document.querySelector('.videos').appendChild(videoEle);
        }

        // Append our tracks if we have a valid stream.
        if(stream) {
            for(const track of stream.getTracks()) {
                peer.addTrack(track, stream);
            }
        }

        // When track are added in the other side, sets the video element src to
        // the stream in use.
        peer.ontrack = function(e) {
            if(!e.streams.length) {
                return;
            }

            videoEle.srcObject = e.streams[0];
        }

        // When trying to find available configuration, just forward the candidate
        // using the signaling channel.
        peer.onicecandidate = function(e) {
            if(!e.candidate) {
                return;
            }

            ws.send(JSON.stringify({
                ice: e.candidate,
                to: id,
            }));
        }

        // Append it to our list of peers for this room.
        peers[id] = peer;

        return peer;
    }

    /**
     * Remove a peer from the collection and close the connection.
     */
    async function removePeer(id) {
        const peer = peers[id];
        await peer.close();
        delete peers[id];

        // Remove the video html element for this user
        findVideoElement(id).remove();
    }
})();
