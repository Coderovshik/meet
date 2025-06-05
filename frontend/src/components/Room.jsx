import React, { useEffect, useRef, useState } from 'react';
import './Room.css';

const Room = ({ username, password }) => {
    const localVideoRef = useRef(null);
    const remoteVideosRef = useRef(null);
    const [localStream, setLocalStream] = useState(null);
    const [connectionStatus, setConnectionStatus] = useState('disconnected');
    const [isAudioEnabled, setIsAudioEnabled] = useState(true);
    const [isVideoEnabled, setIsVideoEnabled] = useState(true);
    const [participantsCount, setParticipantsCount] = useState(0);

    useEffect(() => {
        let pc;
        let ws;

        const init = async () => {
            try {
                setConnectionStatus('connecting');
                const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
                setLocalStream(stream);
                
                if (localVideoRef.current) {
                    localVideoRef.current.srcObject = stream;
                }

                pc = new RTCPeerConnection({
                    iceServers: [
                        { urls: 'stun:stun.l.google.com:19302' },
                        { urls: 'stun:stun1.l.google.com:19302' }
                    ]
                });

                pc.ontrack = event => {
                    if (event.track.kind === 'audio') return;

                    setParticipantsCount(prev => prev + 1);
                    
                    const videoElement = document.createElement('video');
                    videoElement.srcObject = event.streams[0];
                    videoElement.autoplay = true;
                    videoElement.playsInline = true;
                    videoElement.className = 'remote-video';
                    
                    const participantContainer = document.createElement('div');
                    participantContainer.className = 'participant-container';
                    
                    const participantName = document.createElement('div');
                    participantName.className = 'participant-name';
                    participantName.textContent = 'Участник ' + (remoteVideosRef.current?.childElementCount + 1);
                    
                    participantContainer.appendChild(videoElement);
                    participantContainer.appendChild(participantName);
                    
                    remoteVideosRef.current?.appendChild(participantContainer);

                    event.track.onmute = () => {
                        videoElement.play();
                    };

                    event.streams[0].onremovetrack = ({ track }) => {
                        if (track.kind === 'video' && participantContainer.parentNode) {
                            participantContainer.parentNode.removeChild(participantContainer);
                            setParticipantsCount(prev => Math.max(0, prev - 1));
                        }
                    };
                };

                stream.getTracks().forEach(track => pc.addTrack(track, stream));

                ws = new WebSocket(`/ws?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`);

                pc.onicecandidate = e => {
                    if (e.candidate) {
                        ws.send(JSON.stringify({ event: 'candidate', data: JSON.stringify(e.candidate) }));
                    }
                };

                ws.onopen = () => {
                    setConnectionStatus('connected');
                };

                ws.onclose = () => {
                    setConnectionStatus('disconnected');
                    console.log('WebSocket connection closed');
                };

                ws.onerror = (error) => {
                    setConnectionStatus('error');
                    console.error('WebSocket error:', error);
                };

                ws.onmessage = evt => {
                    const msg = JSON.parse(evt.data);
                    if (!msg) return console.log('Failed to parse message');

                    switch (msg.event) {
                        case 'offer':
                            const offer = JSON.parse(msg.data);
                            if (!offer) return console.log('Failed to parse offer');
                            pc.setRemoteDescription(offer);
                            pc.createAnswer().then(answer => {
                                pc.setLocalDescription(answer);
                                ws.send(JSON.stringify({ event: 'answer', data: JSON.stringify(answer) }));
                            });
                            break;

                        case 'candidate':
                            const candidate = JSON.parse(msg.data);
                            if (!candidate) return console.log('Failed to parse candidate');
                            pc.addIceCandidate(candidate);
                            break;

                        default:
                            break;
                    }
                };
            } catch (err) {
                setConnectionStatus('error');
                console.error('Error initializing video call:', err);
            }
        };

        init();

        return () => {
            if (localStream) {
                localStream.getTracks().forEach(track => track.stop());
            }
            if (ws) ws.close();
            if (pc) pc.close();
            setConnectionStatus('disconnected');
        };
    }, [username, password]);

    const toggleAudio = () => {
        if (localStream) {
            const audioTrack = localStream.getAudioTracks()[0];
            if (audioTrack) {
                audioTrack.enabled = !audioTrack.enabled;
                setIsAudioEnabled(audioTrack.enabled);
            }
        }
    };

    const toggleVideo = () => {
        if (localStream) {
            const videoTrack = localStream.getVideoTracks()[0];
            if (videoTrack) {
                videoTrack.enabled = !videoTrack.enabled;
                setIsVideoEnabled(videoTrack.enabled);
            }
        }
    };

    return (
        <div className="room-container">
            <div className="room-header">
                <h2>Видеоконференция</h2>
                <div className={`connection-status status-${connectionStatus}`}>
                    {connectionStatus === 'connected' ? 'Подключено' : 
                     connectionStatus === 'connecting' ? 'Подключение...' : 
                     connectionStatus === 'error' ? 'Ошибка подключения' : 'Отключено'}
                </div>
                <div className="participants-info">
                    <span>Участников: {participantsCount + 1}</span>
                </div>
            </div>
            
            <div className="video-grid">
                <div className="local-video-container">
                    <video 
                        ref={localVideoRef} 
                        autoPlay 
                        muted 
                        playsInline 
                        className={`local-video ${!isVideoEnabled ? 'video-disabled' : ''}`} 
                    />
                    <div className="local-user-info">
                        <span>{username} (Вы)</span>
                    </div>
                    <div className="video-controls">
                        <button 
                            className={`control-button ${!isAudioEnabled ? 'disabled' : ''}`} 
                            onClick={toggleAudio}
                        >
                            {isAudioEnabled ? '🎤' : '🔇'}
                        </button>
                        <button 
                            className={`control-button ${!isVideoEnabled ? 'disabled' : ''}`} 
                            onClick={toggleVideo}
                        >
                            {isVideoEnabled ? '📹' : '📵'}
                        </button>
                    </div>
                </div>
                
                <div className="remote-videos-container" ref={remoteVideosRef}></div>
            </div>
            
            <div className="room-info">
                <h3>Информация о конференции</h3>
                <p>Вы находитесь в общей конференции. Все участники, подключенные к системе, будут видеть и слышать вас.</p>
                <p>Используйте кнопки под вашим видео для управления микрофоном и камерой.</p>
            </div>
        </div>
    );
};

export default Room; 