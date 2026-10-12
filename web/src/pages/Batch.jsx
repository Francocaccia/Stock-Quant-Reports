import { useState } from 'react';
import { API } from '../config';

function Batch({ user }) {
  const [file, setFile] = useState(null);
  const [benchmark, setBenchmark] = useState('SPY');
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState([]);
  const [error, setError] = useState(null);

  const handleFileChange = (e) => {
    if (e.target.files) {
      setFile(e.target.files[0]);
    }
  };

  const handleUpload = async (e) => {
    e.preventDefault();
    if (!file) return;

    setLoading(true);
    setResults([]);
    setError(null);

    const formData = new FormData();
    formData.append('file', file);
    formData.append('benchmark', benchmark);
    formData.append('email', user?.email || '');
    formData.append('analyst', user?.displayName || user?.email || '');

    try {
      const response = await fetch(`${API}/api/batch`, {
        method: 'POST',
        body: formData,
      });

      const data = await response.json();
      if (!response.ok || data.error) {
        throw new Error(data.error || 'Failed to process batch');
      }

      setResults(data.results || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="panel">
      <h2 style={{ color: 'var(--accent-cyan)', marginTop: 0 }}>Batch Processing</h2>
      <p style={{ color: 'var(--text-muted)' }}>Upload a text file containing one Ticker per line to process multiple assets at once.</p>

      <form onSubmit={handleUpload} style={{ marginTop: '30px', maxWidth: '500px' }}>
        <div className="form-group">
          <label className="form-label">Upload File (.txt, .csv)</label>
          <input 
            type="file" 
            accept=".txt,.csv"
            onChange={handleFileChange}
            required
            style={{ color: 'var(--text-main)', marginBottom: '10px' }}
          />
        </div>

        <div className="form-group">
          <label className="form-label">Global Benchmark (Correlation)</label>
          <input 
            type="text" 
            className="input-field" 
            placeholder="default: SPY" 
            value={benchmark}
            onChange={(e) => setBenchmark(e.target.value.toUpperCase())}
          />
        </div>

        <button type="submit" className="btn-primary" disabled={loading || !file} style={{ width: '100%' }}>
          {loading ? 'Crunching Batch...' : 'Run Batch Analysis'}
        </button>
      </form>

      {error && (
        <div style={{ marginTop: '20px', color: 'var(--accent-red)' }}>
          <strong>Error:</strong> {error}
        </div>
      )}

      {results.length > 0 && (
        <div style={{ marginTop: '30px' }}>
          <h3 style={{ borderBottom: '1px solid var(--border-color)', paddingBottom: '10px' }}>Batch Results</h3>
          <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ color: 'var(--text-muted)' }}>
                <th style={{ padding: '10px', borderBottom: '1px solid var(--border-color)' }}>Ticker</th>
                <th style={{ padding: '10px', borderBottom: '1px solid var(--border-color)' }}>Signal</th>
                <th style={{ padding: '10px', borderBottom: '1px solid var(--border-color)' }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {results.map((res, i) => (
                <tr key={i} style={{ borderBottom: '1px solid #222' }}>
                  <td style={{ padding: '10px', fontWeight: 'bold' }}>{res.ticker}</td>
                  <td style={{ padding: '10px', color: res.status === 'Success' ? 'var(--accent-green)' : 'var(--accent-red)' }}>
                    {res.status === 'Success' ? res.signal : 'Failed'}
                  </td>
                  <td style={{ padding: '10px' }}>
                    {res.status === 'Success' ? (
                      <a href={`${API}${res.pdfPath}`} target="_blank" rel="noreferrer" style={{ color: 'var(--accent-cyan)', textDecoration: 'none' }}>
                        View PDF
                      </a>
                    ) : (
                      <span style={{ color: 'var(--accent-red)' }}>{res.error}</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

export default Batch;
