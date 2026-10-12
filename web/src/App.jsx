import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { onAuthStateChanged, signOut } from 'firebase/auth';
import { auth } from './firebase';
import { ADMIN_EMAILS } from './config';
import SingleReport from './pages/SingleReport';
import Batch from './pages/Batch';
import History from './pages/History';
import Admin from './pages/Admin';
import Login from './pages/Login';
import './index.css';

// Placeholder Pages removed

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, (currentUser) => {
      setUser(currentUser);
      setLoading(false);
    });
    return () => unsubscribe();
  }, []);

  const handleLogout = () => {
    signOut(auth);
  };

  if (loading) return <div style={{ color: 'white', padding: '50px', textAlign: 'center' }}>Loading...</div>;

  if (!user) {
    return <Login onLogin={setUser} />;
  }

  // Domain Check (Rough implementation, will refine with Admin Module)
  const isLynchburg = user.email.endsWith('@lynchburg.edu');
  const isAdmin = ADMIN_EMAILS.includes(user.email);

  if (!isLynchburg && !isAdmin) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', backgroundColor: 'var(--bg-dark)' }}>
        <div className="panel" style={{ width: '400px', textAlign: 'center' }}>
          <h2 style={{ color: 'var(--accent-red)' }}>Access Denied</h2>
          <p>Your email <strong>{user.email}</strong> is not authorized to access the Quant Lab.</p>
          <p style={{ color: 'var(--text-muted)' }}>Only @lynchburg.edu accounts or approved guests are allowed.</p>
          <button onClick={handleLogout} className="btn-primary" style={{ marginTop: '20px' }}>Sign Out</button>
        </div>
      </div>
    );
  }

  return (
    <BrowserRouter>
      <div className="app-container">
        {/* Sidebar Navigation */}
        <div className="sidebar">
          <div className="sidebar-title">SMIF QUANT REPORTS</div>
          
          <Link to="/" className="nav-link">Single Report 📊</Link>
          <Link to="/batch" className="nav-link">Batch Upload 📁</Link>
          <Link to="/history" className="nav-link">History 🕒</Link>
          
          <div style={{ flex: 1 }}></div>

          <div style={{ marginBottom: '20px', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Logged in as:<br/>
            <strong style={{ color: isAdmin ? 'var(--accent-magenta)' : 'var(--accent-cyan)' }}>
              {user.email}
            </strong>
          </div>
          
          {isAdmin && (
            <>
              <hr style={{ borderColor: 'var(--border-color)', margin: '15px 0' }} />
              <Link to="/admin" className="nav-link" style={{ color: 'var(--accent-magenta)' }}>
                Admin Panel ⚙️
              </Link>
            </>
          )}

          <button onClick={handleLogout} style={{ marginTop: '10px', background: 'transparent', border: '1px solid var(--accent-red)', color: 'var(--accent-red)', padding: '8px', cursor: 'pointer', borderRadius: '4px' }}>
            Logout
          </button>
        </div>

        {/* Main Workspace */}
        <div className="main-content">
          <Routes>
            <Route path="/" element={<SingleReport user={user} />} />
            <Route path="/batch" element={<Batch user={user} />} />
            <Route path="/history" element={<History user={user} />} />
            {isAdmin && <Route path="/admin" element={<Admin user={user} />} />}
          </Routes>
        </div>
      </div>
    </BrowserRouter>
  );
}

export default App;
