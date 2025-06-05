import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import './Navbar.css';

const Navbar = ({ username, onLogout }) => {
  const location = useLocation();

  return (
    <nav className="navbar">
      <div className="navbar-logo">
        <Link to="/room">Meet</Link>
      </div>
      
      <div className="navbar-user">
        <span className="username">{username}</span>
      </div>
      
      <div className="navbar-menu">
        <Link 
          to="/room" 
          className={location.pathname === '/room' ? 'active' : ''}
        >
          Видеоконференция
        </Link>
        <Link 
          to="/logs" 
          className={location.pathname === '/logs' ? 'active' : ''}
        >
          Логи
        </Link>
        <button className="logout-button" onClick={onLogout}>
          Выход
        </button>
      </div>
    </nav>
  );
};

export default Navbar; 