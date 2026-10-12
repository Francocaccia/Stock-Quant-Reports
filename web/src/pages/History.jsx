import { useState, useEffect } from 'react';

import { API } from '../config';

const SIGNAL_COLOR = (s) => ({
  bg:   s?.includes('BUY') ? '#1b4332' : s?.includes('SELL') ? '#432818' : '#2a2a2a',
  text: s?.includes('BUY') ? '#4ade80' : s?.includes('SELL') ? '#fb923c' : '#aaa',
});

function BarRow({ label, count, max, color }) {
  const pct = max > 0 ? (count / max) * 100 : 0;
  return (
    <div style={{ marginBottom: 10 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.82rem', marginBottom: 4 }}>
        <span style={{ fontWeight: 600 }}>{label}</span>
        <span style={{ color: 'var(--text-muted)' }}>{count}</span>
      </div>
      <div style={{ background: '#1a1a1a', borderRadius: 4, height: 8 }}>
        <div style={{ width: `${pct}%`, background: color, borderRadius: 4, height: '100%', transition: 'width 0.5s ease' }} />
      </div>
    </div>
  );
}

function Tab({ label, active, onClick }) {
  return (
    <button onClick={onClick} style={{
      padding: '8px 18px', borderRadius: 8, border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: '0.85rem',
      background: active ? 'var(--accent-cyan)' : '#1a1a1a',
      color: active ? '#000' : 'var(--text-muted)',
      transition: 'all 0.2s',
    }}>{label}</button>
  );
}

function History({ user }) {
  const [reports, setReports] = useState([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState('reports');

  // Filters
  const [fTicker, setFTicker] = useState('');
  const [fSignal, setFSignal] = useState('');
  const [fFrom,   setFFrom]   = useState('');
  const [fTo,     setFTo]     = useState('');

  useEffect(() => {
    if (user?.email) fetchReports();
  }, [user]);

  const fetchReports = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({ email: user.email });
      const res  = await fetch(`${API}/api/history?${params}`);
      const data = await res.json();
      setReports(Array.isArray(data) ? data : []);
    } catch (e) {
      console.error(e);
      setReports([]);
    } finally {
      setLoading(false);
    }
  };

  // Derived metrics
  const buys   = reports.filter(r => r.signal?.includes('BUY')).length;
  const sells  = reports.filter(r => r.signal?.includes('SELL')).length;
  const holds  = reports.filter(r => r.signal?.includes('HOLD')).length;

  const tickerMap = {};
  reports.forEach(r => { tickerMap[r.ticker] = (tickerMap[r.ticker] || 0) + 1; });
  const topTickers = Object.entries(tickerMap).sort((a,b) => b[1]-a[1]).slice(0, 10);
  const maxTicker  = topTickers[0]?.[1] || 1;

  // Filtered rows (client-side for snappy UX)
  const filtered = reports.filter(r => {
    if (fTicker && !r.ticker?.toUpperCase().includes(fTicker.toUpperCase())) return false;
    if (fSignal && !r.signal?.toUpperCase().includes(fSignal.toUpperCase())) return false;
    if (fFrom && new Date(r.ts) < new Date(fFrom)) return false;
    if (fTo   && new Date(r.ts) > new Date(fTo + 'T23:59:59')) return false;
    return true;
  });

  const exportCSV = () => {
    const headers = ['Timestamp','Analyst','Ticker','Signal','Source','PDF'];
    const rows = filtered.map(r => [
      new Date(r.ts).toLocaleString(), r.analyst, r.ticker, r.signal,
      r.source || '', `${API}${r.pdf_file || ''}`
    ]);
    const csv = [headers, ...rows].map(row => row.map(c => `"${c}"`).join(',')).join('\n');
    const link = document.createElement('a');
    link.href = URL.createObjectURL(new Blob([csv], { type: 'text/csv' }));
    link.download = `my_history_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
  };

  const inputStyle = { background:'#111', border:'1px solid #333', color:'#fff', padding:'7px 10px', borderRadius:6, fontSize:'0.85rem' };

  if (loading) return (
    <div className="panel" style={{ textAlign:'center', paddingTop:60 }}>
      <div className="loading-spinner" />
      <p style={{ marginTop:16, color:'var(--text-muted)' }}>Loading your reports…</p>
    </div>
  );

  return (
    <div className="panel">
      {/* Header */}
      <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:16 }}>
        <h2 style={{ color:'var(--accent-cyan)', margin:0 }}>Report History</h2>
        <div style={{ display:'flex', gap: 10 }}>
          <button onClick={exportCSV} style={{ padding:'7px 14px', background:'#111', border:'1px solid #00f0ff', color:'#00f0ff', borderRadius:7, cursor:'pointer', fontSize:'0.8rem', fontWeight:600 }}>📥 Export CSV</button>
          <button onClick={fetchReports} style={{ padding:'7px 14px', background:'#222', border:'1px solid #444', color:'#aaa', borderRadius:7, cursor:'pointer', fontSize:'0.8rem' }}>↻ Refresh</button>
        </div>
      </div>

      {/* KPI row */}
      <div style={{ display:'flex', gap:12, flexWrap:'wrap', marginBottom:20 }}>
        {[
          { label:'My Reports', value: reports.length, color:'var(--accent-cyan)' },
          { label:'BUY',  value: buys,  color:'#4ade80' },
          { label:'SELL', value: sells, color:'#fb923c' },
          { label:'HOLD', value: holds, color:'#aaa'    },
        ].map(k => (
          <div key={k.label} style={{ flex:1, minWidth:90, padding:'12px 16px', background:'#0d0d0d', border:'1px solid #2a2a2a', borderRadius:10 }}>
            <div style={{ fontSize:'0.7rem', color:'var(--text-muted)', textTransform:'uppercase', letterSpacing:1 }}>{k.label}</div>
            <div style={{ fontSize:'1.6rem', fontWeight:800, color:k.color }}>{k.value}</div>
          </div>
        ))}
      </div>

      {/* Tabs */}
      <div style={{ display:'flex', gap:8, marginBottom:20 }}>
        <Tab label="📋 Reports" active={tab==='reports'} onClick={() => setTab('reports')} />
        <Tab label="📊 My Metrics" active={tab==='metrics'} onClick={() => setTab('metrics')} />
      </div>

      {/* ── TAB: REPORTS ─────────────────────────────────────────────────────── */}
      {tab === 'reports' && (
        <>
          <div style={{ display:'grid', gridTemplateColumns:'repeat(auto-fit,minmax(140px,1fr))', gap:10, marginBottom:12 }}>
            <input placeholder="Ticker" value={fTicker} onChange={e=>setFTicker(e.target.value)} style={inputStyle} />
            <select value={fSignal} onChange={e=>setFSignal(e.target.value)} style={inputStyle}>
              <option value="">All Signals</option>
              <option value="BUY">BUY</option>
              <option value="SELL">SELL</option>
              <option value="HOLD">HOLD</option>
            </select>
            <input type="date" value={fFrom} onChange={e=>setFFrom(e.target.value)} style={inputStyle} />
            <input type="date" value={fTo}   onChange={e=>setFTo(e.target.value)}   style={inputStyle} />
            <button onClick={()=>{setFTicker('');setFSignal('');setFFrom('');setFTo('');}}
              style={{ background:'transparent', border:'1px solid #444', color:'#aaa', borderRadius:6, cursor:'pointer', fontSize:'0.82rem' }}>
              Clear
            </button>
          </div>

          <p style={{ fontSize:'0.8rem', color:'var(--text-muted)', marginBottom:10 }}>
            Showing {filtered.length} of {reports.length} records
          </p>

          {filtered.length === 0 ? (
            <p style={{ textAlign:'center', color:'var(--text-muted)', marginTop:30 }}>No reports found matching the current filters.</p>
          ) : (
            <div style={{ overflowX:'auto' }}>
              <table style={{ width:'100%', borderCollapse:'collapse', textAlign:'left', fontSize:'0.85rem' }}>
                <thead>
                  <tr style={{ borderBottom:'2px solid #333', color:'var(--accent-magenta)' }}>
                    {['Date','Ticker','Signal','PDF'].map(h => <th key={h} style={{ padding:'10px 12px' }}>{h}</th>)}
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((r, i) => {
                    const sc = SIGNAL_COLOR(r.signal);
                    return (
                      <tr key={i} className="table-row-hover" style={{ borderBottom:'1px solid #1a1a1a' }}>
                        <td style={{ padding:'10px 12px', color:'var(--text-muted)', fontSize:'0.82rem' }}>{new Date(r.ts).toLocaleString()}</td>
                        <td style={{ padding:'10px 12px', fontWeight:700 }}>{r.ticker}</td>
                        <td style={{ padding:'10px 12px' }}>
                          <span style={{ padding:'3px 8px', borderRadius:4, background:sc.bg, color:sc.text, fontWeight:700, fontSize:'0.78rem' }}>{r.signal}</span>
                        </td>
                        <td style={{ padding:'10px 12px' }}>
                          <a href={`${API}${r.pdf_file}`} target="_blank" rel="noreferrer"
                            style={{ color:'var(--accent-cyan)', textDecoration:'none', fontWeight:600 }}>Download ↗</a>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {/* ── TAB: METRICS ─────────────────────────────────────────────────────── */}
      {tab === 'metrics' && (
        reports.length === 0 ? (
          <p style={{ textAlign:'center', color:'var(--text-muted)', marginTop:30 }}>Generate your first report to see your metrics.</p>
        ) : (
          <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:20 }}>

            <div style={{ padding:20, background:'#0d0d0d', border:'1px solid #2a2a2a', borderRadius:10 }}>
              <h3 style={{ color:'var(--accent-cyan)', margin:'0 0 16px' }}>My Signal Distribution</h3>
              <BarRow label="BUY"  count={buys}  max={reports.length} color="#4ade80" />
              <BarRow label="SELL" count={sells} max={reports.length} color="#fb923c" />
              <BarRow label="HOLD" count={holds} max={reports.length} color="#6b7280" />
              <p style={{ fontSize:'0.75rem', color:'var(--text-muted)', marginTop:12 }}>
                My bullish rate: <strong style={{ color:'#4ade80' }}>
                  {((buys / reports.length) * 100).toFixed(1)}%
                </strong>
              </p>
            </div>

            <div style={{ padding:20, background:'#0d0d0d', border:'1px solid #2a2a2a', borderRadius:10 }}>
              <h3 style={{ color:'var(--accent-cyan)', margin:'0 0 16px' }}>My Top Analyzed Tickers</h3>
              {topTickers.map(([ticker, count]) => (
                <BarRow key={ticker} label={ticker} count={count} max={maxTicker} color="var(--accent-magenta)" />
              ))}
            </div>

          </div>
        )
      )}
    </div>
  );
}

export default History;
