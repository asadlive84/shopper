// src/pages/Home.tsx
import React, { useEffect, useState } from 'react';

const Home = () => {
  const [user, setUser] = useState({ name: '', email: '' });

  useEffect(() => {
    // Retrieve user data from localStorage
    const userData = localStorage.getItem('user');
    if (userData) {
      setUser(JSON.parse(userData));
    }
  }, []);

  return (
    <div
      style={{
        padding: '2rem',
        textAlign: 'center',
      }}
    >
      <h1 style={{ marginBottom: '1rem', color: '#333' }}>Welcome to the Home Page</h1>
      <div>
        <p style={{ fontSize: '1.2rem', color: '#555' }}>Name: {user.name}</p>
        <p style={{ fontSize: '1.2rem', color: '#555' }}>Email: {user.email}</p>
      </div>
    </div>
  );
};

export default Home;