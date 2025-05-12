import React, { useState, useEffect, useRef } from 'react';

function App() {
  const [username, setUsername] = useState(localStorage.getItem('username') || '');
  const [password, setPassword] = useState(localStorage.getItem('password') || '');
  const [loggedIn, setLoggedIn] = useState(!!username && !!password);
  const [roomId, setRoomId] = useState('');
  const [rooms, setRooms] = useState([]);
  const [connected, setConnected] = useState(false);
  const localVideoRef = useRef(null);
  const remoteVideoRef = useRef(null);
  const pcRef = useRef(null);
  const wsRef = useRef(null);

  const getAuthHeader = () => {
    return 'Basic ' + btoa(username + ':' + password);
  };

  const fetchRooms = async () => {
    try {
      const response = await fetch('/api/rooms', {
        headers: { Authorization: getAuthHeader() }
      });
      const data = await response.json();
      setRooms(data);
    } catch { }
  };

  useEffect(() => {
    if (loggedIn) fetchRooms();
  }, [loggedIn]);

  const register = async () => {
    try {
      await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      alert('User registered. Now you can login.');
    } catch {
      alert('Registration failed');
    }
  };

  const login = async () => {
    try {
      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      if (resp.status === 200) {
        localStorage.setItem('username', username);
        localStorage.setItem('password', password);
        setLoggedIn(true);
      } else {
        alert('Неверный логин или пароль');
      }
    } catch {
      alert('Ошибка сервера');
    }
  };

  const logout = () => {
    localStorage.removeItem('username');
    localStorage.removeItem('password');
    setUsername('');
    setPassword('');
    setLoggedIn(false);
  };

  const createRoom = async () => {
    if (!roomId) return;
    await fetch('/api/rooms', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: getAuthHeader()
      },
      body: JSON.stringify({ ID: roomId }),
    });
    setRoomId('');
    await fetchRooms();
  };

  const connectToRoom = async (room) => {
    if (connected) return;
    const ws = new WebSocket(
      `ws://localhost:8080/ws?room=${encodeURIComponent(room)}&username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`
    );
    wsRef.current = ws;

    ws.onopen = () => {
      const pc = new RTCPeerConnection();
      pcRef.current = pc;

      pc.onicecandidate = (e) => {
        if (e.candidate) {
          ws.send(JSON.stringify({
            type: 'candidate',
            candidate: e.candidate.candidate,
          }));
        }
      };

      pc.ontrack = (e) => {
        remoteVideoRef.current.srcObject = e.streams[0];
      };

      navigator.mediaDevices.getUserMedia({ video: true, audio: true }).then(stream => {
        localVideoRef.current.srcObject = stream;
        stream.getTracks().forEach((track) => pc.addTrack(track, stream));

        pc.createOffer().then(offer => {
          pc.setLocalDescription(offer);
          ws.send(JSON.stringify({
            type: 'offer',
            sdp: offer.sdp,
          }));
        });
      });
    };

    ws.onmessage = async (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === 'answer') {
        await pcRef.current.setRemoteDescription(new RTCSessionDescription({
          type: 'answer',
          sdp: msg.sdp,
        }));
      }
    };

    setConnected(true);
  };

  if (!loggedIn) {
    return (
      <div>
        <h1>Login / Register</h1>
        <input type="text" placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} /><br />
        <input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} /><br />
        <button onClick={login}>Login</button>
        <button onClick={register}>Register</button>
      </div>
    );
  }

  return (
    <div>
      <h1>SFU Webinar Mode (User: {username})</h1>
      <button onClick={logout}>Logout</button><br /><br />
      <input type="text" placeholder="Room ID" value={roomId} onChange={(e) => setRoomId(e.target.value)} />
      <button onClick={createRoom}>Create Room</button>
      <h2>Available Rooms</h2>
      <ul>
        {rooms.map((room) => (
          <li key={room}>
            {room}{' '}
            <button onClick={() => connectToRoom(room)} disabled={connected}>
              Connect
            </button>
          </li>
        ))}
      </ul>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <div style={{ marginBottom: '10px' }}>
          <h3>Вы (локальное видео)</h3>
          <video ref={localVideoRef} autoPlay muted playsInline style={{ width: '300px', border: '1px solid black' }} />
        </div>
        <div>
          <h3>Ведущий (удалённое видео)</h3>
          <video ref={remoteVideoRef} autoPlay playsInline style={{ width: '300px', border: '1px solid black' }} />
        </div>
      </div>
    </div>
  );
}

export default App;
