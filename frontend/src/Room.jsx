import React, { useEffect, useRef } from 'react';

const Room = ({ username, password }) => {
  const localVideoRef = useRef(null);
  const remoteVideosRef = useRef(null);

  useEffect(() => {
    let pc;
    let ws;

    navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      .then(stream => {
        pc = new RTCPeerConnection();

        pc.ontrack = event => {
          if (event.track.kind === 'audio') return;

          const el = document.createElement(event.track.kind);
          el.srcObject = event.streams[0];
          el.autoplay = true;
          el.controls = true;

          remoteVideosRef.current?.appendChild(el);

          event.track.onmute = () => {
            el.play();
          };

          event.streams[0].onremovetrack = ({ track }) => {
            if (el.parentNode) {
              el.parentNode.removeChild(el);
            }
          };
        };

        if (localVideoRef.current) {
          localVideoRef.current.srcObject = stream;
        }

        stream.getTracks().forEach(track => pc.addTrack(track, stream));

        ws = new WebSocket(`wss://amogus.root-hub.ru/ws?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`);

        pc.onicecandidate = e => {
          if (e.candidate) {
            ws.send(JSON.stringify({ event: 'candidate', data: JSON.stringify(e.candidate) }));
          }
        };

        ws.onclose = () => {
          window.alert('WebSocket has closed');
        };

        ws.onmessage = evt => {
          const msg = JSON.parse(evt.data);
          if (!msg) return console.log('failed to parse msg');

          switch (msg.event) {
            case 'offer':
              const offer = JSON.parse(msg.data);
              if (!offer) return console.log('failed to parse offer');
              pc.setRemoteDescription(offer);
              pc.createAnswer().then(answer => {
                pc.setLocalDescription(answer);
                ws.send(JSON.stringify({ event: 'answer', data: JSON.stringify(answer) }));
              });
              break;

            case 'candidate':
              const candidate = JSON.parse(msg.data);
              if (!candidate) return console.log('failed to parse candidate');
              pc.addIceCandidate(candidate);
              break;

            default:
              break;
          }
        };

        ws.onerror = evt => {
          console.log('ERROR: ' + evt.data);
        };
      })
      .catch(err => window.alert(err));

    return () => {
      if (ws) ws.close();
      if (pc) pc.close();
    };
  }, [username, password]);

  return (
    <div>
      <h3>Local Video</h3>
      <video ref={localVideoRef} width="160" height="120" autoPlay muted playsInline />

      <h3>Remote Video</h3>
      <div ref={remoteVideosRef}></div>

      <h3>Logs</h3>
      <div id="logs"></div>
    </div>
  );
};

export default Room;