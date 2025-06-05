import React, { useState, useEffect } from 'react';
import './UserLogs.css';

const UserLogs = ({ username, password }) => {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [limit, setLimit] = useState(50);

  useEffect(() => {
    fetchLogs();
  }, [limit]);

  const fetchLogs = async () => {
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch(`/api/logs?limit=${limit}`, {
        headers: {
          'Authorization': `Basic ${username}:${password}`
        }
      });
      
      if (!response.ok) {
        throw new Error(`Ошибка: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      setLogs(data);
    } catch (err) {
      console.error('Ошибка при получении логов:', err);
      setError('Не удалось загрузить логи. Пожалуйста, попробуйте позже.');
    } finally {
      setLoading(false);
    }
  };

  // Форматирование даты и времени
  const formatDateTime = (timestamp) => {
    const date = new Date(timestamp);
    return new Intl.DateTimeFormat('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    }).format(date);
  };

  // Преобразование типа действия в читаемый текст
  const getActionText = (action) => {
    const actionMap = {
      'registration': 'Регистрация',
      'login': 'Вход в систему',
      'login_failed': 'Неудачная попытка входа',
      'room_connection': 'Подключение к комнате',
      'room_disconnection': 'Отключение от комнаты',
      'room_connection_failed': 'Неудачная попытка подключения к комнате',
      'add_track': 'Добавление аудио/видео потока'
    };
    
    return actionMap[action] || action;
  };

  // Получение соответствующего класса для типа действия
  const getActionClass = (action) => {
    if (action.includes('failed')) {
      return 'action-failed';
    } else if (action === 'login' || action === 'registration') {
      return 'action-auth';
    } else if (action.includes('room')) {
      return 'action-room';
    } else if (action.includes('track')) {
      return 'action-track';
    }
    return '';
  };

  const handleRefresh = () => {
    fetchLogs();
  };

  return (
    <div className="logs-container">
      <div className="logs-header">
        <h2>Логи активности пользователя</h2>
        <div className="logs-controls">
          <div className="limit-control">
            <label htmlFor="limit">Показать записей:</label>
            <select 
              id="limit" 
              value={limit} 
              onChange={(e) => setLimit(Number(e.target.value))}
            >
              <option value={10}>10</option>
              <option value={20}>20</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
          </div>
          <button className="refresh-button" onClick={handleRefresh} disabled={loading}>
            {loading ? 'Загрузка...' : 'Обновить'}
          </button>
        </div>
      </div>
      
      {error && <div className="logs-error">{error}</div>}
      
      {loading ? (
        <div className="logs-loading">Загрузка логов...</div>
      ) : logs.length === 0 ? (
        <div className="logs-empty">Нет доступных логов</div>
      ) : (
        <div className="logs-table-container">
          <table className="logs-table">
            <thead>
              <tr>
                <th>Дата и время</th>
                <th>Действие</th>
                <th>Детали</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log, index) => (
                <tr key={index} className={getActionClass(log.action)}>
                  <td className="timestamp">{formatDateTime(log.timestamp)}</td>
                  <td className="action">{getActionText(log.action)}</td>
                  <td className="details">{log.details}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default UserLogs; 