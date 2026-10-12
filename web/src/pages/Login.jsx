import { useState } from 'react';
import { signInWithPopup, GoogleAuthProvider } from 'firebase/auth';
import { auth } from '../firebase';

function Login({ onLogin }) {
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleGoogleLogin = async () => {
    setLoading(true);
    setError('');
    
    try {
      const provider = new GoogleAuthProvider();
      // Force account selection to avoid auto-login with wrong accounts
      provider.setCustomParameters({ prompt: 'select_account' });
      
      const result = await signInWithPopup(auth, provider);
      
      // We will check the domain here later (@lynchburg.edu)
      onLogin(result.user);
      
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', backgroundColor: 'var(--bg-dark)' }}>
      <div className="panel" style={{ width: '400px', textAlign: 'center' }}>
        <h1 style={{ color: 'var(--accent-cyan)', marginBottom: '10px' }}>SMIF QUANT REPORTS</h1>
        <p style={{ color: 'var(--text-muted)', marginBottom: '40px' }}>Institutional Research Portal — Lynchburg College</p>
        
        <button 
          onClick={handleGoogleLogin} 
          disabled={loading}
          className="btn-primary" 
          style={{ width: '100%', padding: '15px', fontSize: '1.1rem', backgroundColor: '#ffffff', color: '#000000', border: 'none' }}>
            <img src="https://developers.google.com/identity/images/g-logo.png" alt="Google" style={{ width: '20px', verticalAlign: 'middle', marginRight: '10px' }}/>
            {loading ? 'Authenticating...' : 'Sign in with Google'}
        </button>

        {error && <p style={{ color: 'var(--accent-red)', marginTop: '20px' }}>{error}</p>}
      </div>
    </div>
  );
}

export default Login;
