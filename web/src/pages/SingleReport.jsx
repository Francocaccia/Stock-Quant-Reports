import { useState } from 'react';
import { API } from '../config';

function SingleReport({ user }) {
  const [ticker, setTicker] = useState('');
  const [benchmark, setBenchmark] = useState('SPY');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);

  const handleGenerate = async (e) => {
    e.preventDefault();
    setLoading(true);
    setResult(null);

    try {
      // Pointing to Go backend on port 8080
      const response = await fetch(`${API}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ticker, benchmark, email: user?.email, analyst: user?.displayName || user?.email })
      });
      
      const data = await response.json();
      
      if (!response.ok || data.error) {
        throw new Error(data.error || 'Failed to generate report');
      }

      setResult({
        pdfUrl: `${API}${data.pdfPath}`,
        signal: data.signal
      });
    } catch (err) {
      alert(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="panel">
      <h2 style={{ color: 'var(--accent-cyan)', marginTop: 0 }}>Analyze Asset</h2>
      <p style={{ color: 'var(--text-muted)' }}>Generate a comprehensive multi-timeframe PDF report.</p>

      <form onSubmit={handleGenerate} style={{ marginTop: '30px', maxWidth: '400px' }}>
        <div className="form-group">
          <label className="form-label">Ticker (Asset)</label>
          <input 
            type="text" 
            className="input-field" 
            placeholder="e.g. NVDA, AAPL, MELI" 
            value={ticker}
            onChange={(e) => setTicker(e.target.value.toUpperCase())}
            required
          />
        </div>

        <div className="form-group">
          <label className="form-label">Benchmark (Correlation)</label>
          <input 
            type="text" 
            className="input-field" 
            placeholder="default: SPY" 
            value={benchmark}
            onChange={(e) => setBenchmark(e.target.value.toUpperCase())}
          />
        </div>

        <button type="submit" className="btn-primary" disabled={loading} style={{ width: '100%' }}>
          {loading ? 'Crunching Numbers...' : 'Generate Report'}
        </button>
      </form>

      {result && (
        <div style={{ marginTop: '30px', padding: '20px', border: '1px solid var(--accent-magenta)', borderRadius: '4px' }}>
          <h3 style={{ marginTop: 0 }}>Analysis Complete</h3>
          <p><strong>Final Signal:</strong> {result.signal}</p>
          <a href={result.pdfUrl} target="_blank" rel="noreferrer" className="btn-primary" style={{ display: 'inline-block', textDecoration: 'none', marginTop: '10px' }}>
            Open PDF Report
          </a>
        </div>
      )}
    </div>
  );
}

export default SingleReport;
