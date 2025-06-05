import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Login from './components/Login';
import Register from './components/Register';
import Room from './components/Room';
import UserLogs from './components/UserLogs';
import Navbar from './components/Navbar';
import './App.css';

function App() {
  const [username, setUsername] = useState(localStorage.getItem('username') || '');
  const [password, setPassword] = useState(localStorage.getItem('password') || '');
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Проверяем аутентификацию при загрузке
    const checkAuth = async () => {
      if (username && password) {
        try {
          const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
          });

          setIsAuthenticated(response.status === 200);
        } catch (error) {
          console.error('Auth check error:', error);
          setIsAuthenticated(false);
        }
      } else {
        setIsAuthenticated(false);
      }
      setLoading(false);
    };

    checkAuth();
  }, [username, password]);

  const handleLogin = (username, password) => {
    setUsername(username);
    setPassword(password);
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    localStorage.removeItem('username');
    localStorage.removeItem('password');
    setUsername('');
    setPassword('');
    setIsAuthenticated(false);
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>Загрузка...</p>
      </div>
    );
  }

  return (
    <Router>
      <div className="app-container">
        {isAuthenticated && (
          <Navbar username={username} onLogout={handleLogout} />
        )}
        
        <main className="main-content">
          <Routes>
            <Route 
              path="/login" 
              element={
                isAuthenticated ? 
                <Navigate to="/room" replace /> : 
                <Login onLogin={handleLogin} />
              } 
            />
            
            <Route 
              path="/register" 
              element={
                isAuthenticated ? 
                <Navigate to="/room" replace /> : 
                <Register />
              } 
            />
            
            <Route 
              path="/room" 
              element={
                isAuthenticated ? 
                <Room username={username} password={password} /> : 
                <Navigate to="/login" replace />
              } 
            />
            
            <Route 
              path="/logs" 
              element={
                isAuthenticated ? 
                <UserLogs username={username} password={password} /> : 
                <Navigate to="/login" replace />
              } 
            />
            
            <Route 
              path="*" 
              element={
                isAuthenticated ? 
                <Navigate to="/room" replace /> : 
                <Navigate to="/login" replace />
              } 
            />
          </Routes>
        </main>
        
        <footer className="app-footer">
          <p>&copy; {new Date().getFullYear()} Meet - Сервис видеоконференций</p>
        </footer>
      </div>
    </Router>
  );
}

export default App;
