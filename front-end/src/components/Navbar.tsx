import React, { useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import socket from '../utils/socket';

const Navbar = () => {
  const { isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    // Handle incoming messages from the server
    socket.onMessage((event:any) => {
      const message = JSON.parse(event.data);
      console.log('Message from server:', message);
    });

    // কম্পোনেন্ট আনমাউন্টে কানেকশন বন্ধ করার দরকার নেই, কারণ এটি সিঙ্গলটন
    // return () => socket.close();
  }, []);

  const handleLogout = () => {
    const user = JSON.parse(localStorage.getItem('user') || '{}');

    // ব্যাকএন্ডের ফরম্যাটে মেসেজ পাঠানো
    const message = JSON.stringify({
      type: 'logout', // 'event' এর পরিবর্তে 'type'
      data: {
        user_id: user.id, // 'userId' এর পরিবর্তে 'user_id'
        email: user.email, // অতিরিক্ত ফিল্ড, ব্যাকএন্ডে হ্যান্ডল করতে হবে
      },
    });
    socket.send(message);

    // লোকাল স্টোরেজ থেকে ডাটা মুছে ফেলা এবং স্টেট আপডেট
    logout();

    // লগইন পেজে রিডিরেক্ট
    navigate('/login');
  };

  return (
    <div
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 1000,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '1rem 2rem',
        backgroundColor: '#ffffff',
        boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)',
      }}
    >
      <div style={{ fontSize: '1.5rem', fontWeight: 'bold', color: '#007bff' }}>MyApp</div>
      <div style={{ display: 'flex', gap: '1.5rem' }}>
        <Link to="/" style={{ textDecoration: 'none', color: '#333', fontSize: '1rem', transition: 'color 0.3s ease' }}>
          Home
        </Link>
        {isAuthenticated && (
          <Link
            to="/profile"
            style={{ textDecoration: 'none', color: '#333', fontSize: '1rem', transition: 'color 0.3s ease' }}
          >
            Profile
          </Link>
        )}
        {isAuthenticated && (
          <Link
            to="/product"
            style={{ textDecoration: 'none', color: '#333', fontSize: '1rem', transition: 'color 0.3s ease' }}
          >
            Product
          </Link>
        )}
        {!isAuthenticated && (
          <Link
            to="/login"
            style={{ textDecoration: 'none', color: '#333', fontSize: '1rem', transition: 'color 0.3s ease' }}
          >
            Login
          </Link>
        )}
        {!isAuthenticated && (
          <Link
            to="/signup"
            style={{ textDecoration: 'none', color: '#333', fontSize: '1rem', transition: 'color 0.3s ease' }}
          >
            Signup
          </Link>
        )}
        {isAuthenticated && (
          <button
            onClick={handleLogout}
            style={{
              background: 'none',
              border: 'none',
              color: '#333',
              fontSize: '1rem',
              cursor: 'pointer',
              transition: 'color 0.3s ease',
            }}
          >
            Logout
          </button>
        )}
      </div>
    </div>
  );
};

export default Navbar;