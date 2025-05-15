import React, { useState } from 'react';
import Room from './Room';

function App() {
  const [username, setUsername] = useState(localStorage.getItem('username') || '');
  const [password, setPassword] = useState(localStorage.getItem('password') || '');
  const [loggedIn, setLoggedIn] = useState(!!username && !!password);

  const register = async () => {
    try {
      await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (resp.status !== 201) {
        alert('Registration failed');
        return;
      }

      alert('User registered. Now you can login.');
    } catch {
      alert('Registration error');
    }
  };

  const login = async () => {
    try {
      const resp = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (resp.status !== 200) {
        alert('Login failed');
        return;
      }

      localStorage.setItem('username', username);
      localStorage.setItem('password', password);
      setLoggedIn(true);
    } catch {
      alert('Login error');
    }
  };

  const logout = () => {
    localStorage.removeItem('username');
    localStorage.removeItem('password');
    setUsername('');
    setPassword('');
    setLoggedIn(false);
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
      <h1>Meet (User: {username})</h1>
      <button onClick={logout}>Logout</button><br /><br />
      <Room username={username} password={password} />
    </div>
  );
}

export default App;
