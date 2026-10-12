import { useState, useEffect } from 'react';

import { API } from '../config';

const SIGNAL_COLOR = (s) => ({
  bg:   s?.includes('BUY') ? '#1b4332' : s?.includes('SELL') ? '#432818' : '#2a2a2a',
  text: s?.includes('BUY') ? '#4ade80' : s?.includes('SELL') ? '#fb923c' : '#aaa',
});

// ─── Stat Card ───────────────────────────────────────────────────────────────
function StatCard({ label, value, color = 'var(--accent-cyan)' }) {
  return (
    <div style={{ flex: 1, minWidth: 120, padding: '16px 20px', background: '#0d0d0d', border: '1px solid #2a2a2a', borderRadius: 10 }}>
      <div style={{ fontSize: '0.72rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: 1 }}>{label}</div>
      <div style={{ fontSize: '1.8rem', fontWeight: 800, color, marginTop: 4 }}>{value}</div>
    </div>
  );
}

// ─── Simple horizontal bar chart ─────────────────────────────────────────────
function BarRow({ label, count, max, color = 'var(--accent-cyan)' }) {
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

// ─── Tab button ──────────────────────────────────────────────────────────────
function Tab({ label, active, onClick }) {
  return (
    <button onClick={onClick} style={{
      padding: '8px 18px', borderRadius: 8, border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: '0.85rem',
      background: active ? 'var(--accent-magenta)' : '#1a1a1a',
      color: active ? '#fff' : 'var(--text-muted)',
      transition: 'all 0.2s',
    }}>{label}</button>
  );
}

// ─── Main Admin Component ─────────────────────────────────────────────────────
function Admin() {
  const [reports, setReports]   = useState([]);
  const [loading, setLoading]   = useState(true);
  const [tab, setTab]           = useState('reports');
  const [selectedUser, setSelectedUser] = useState(null); // for Users drill-down

  // Filters (Reports tab)
  const [fTicker, setFTicker] = useState('');
  const [fSignal, setFSignal] = useState('');
  const [fFrom,   setFFrom]   = useState('');
  const [fTo,     setFTo]     = useState('');

  useEffect(() => { fetchAll(); }, []);

  const fetchAll = async () => {
    setLoading(true);
    try {
      const res  = await fetch(`${API}/api/admin/history`);
      const data = await res.json();
      setReports(Array.isArray(data) ? data : []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  // ── derived data ────────────────────────────────────────────────────────────
  const filtered = reports.filter(r => {
    if (fTicker && !r.ticker?.toUpperCase().includes(fTicker.toUpperCase())) return false;
    if (fSignal && !r.signal?.toUpperCase().includes(fSignal.toUpperCase())) return false;
    if (fFrom && new Date(r.ts) < new Date(fFrom)) return false;
    if (fTo   && new Date(r.ts) > new Date(fTo + 'T23:59:59')) return false;
    return true;
  });

  const analystMap = {};
  reports.forEach(r => {
    if (!analystMap[r.analyst]) analystMap[r.analyst] = { total: 0, buy: 0, sell: 0, hold: 0 };
    analystMap[r.analyst].total++;
    if (r.signal?.includes('BUY'))  analystMap[r.analyst].buy++;
    if (r.signal?.includes('SELL')) analystMap[r.analyst].sell++;
    if (r.signal?.includes('HOLD')) analystMap[r.analyst].hold++;
  });

  const tickerMap = {};
  reports.forEach(r => { tickerMap[r.ticker] = (tickerMap[r.ticker] || 0) + 1; });
  const topTickers = Object.entries(tickerMap).sort((a,b) => b[1]-a[1]).slice(0, 10);
  const maxTicker  = topTickers[0]?.[1] || 1;

  const buys  = reports.filter(r => r.signal?.includes('BUY')).length;
  const sells = reports.filter(r => r.signal?.includes('SELL')).length;
  const holds = reports.filter(r => r.signal?.includes('HOLD')).length;

  const exportCSV = () => {
    const headers = ['Timestamp','Analyst','Ticker','Signal','Source','PDF'];
    const rows = filtered.map(r => [
      new Date(r.ts).toLocaleString(), r.analyst, r.ticker, r.signal,
      r.source || '', `${API}${r.pdf_file || ''}`
    ]);
    const csv = [headers, ...rows].map(row => row.map(c => `"${c}"`).join(',')).join('\n');
    const link = document.createElement('a');
    link.href = URL.createObjectURL(new Blob([csv], { type: 'text/csv' }));
    link.download = `quant_admin_${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
  };

  if (loading) return (
    <div className="panel" style={{ textAlign: 'center', paddingTop: 60 }}>
      <div className="loading-spinner" />
      <p style={{ marginTop: 16, color: 'var(--text-muted)' }}>Loading admin data…</p>
    </div>
  );

  return (
    <div className="panel">
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <h2 style={{ color: 'var(--accent-magenta)', margin: 0 }}>Admin Panel ⚙️</h2>
        <button onClick={fetchAll} style={{ padding: '7px 14px', background: '#222', border: '1px solid #444', color: '#aaa', borderRadius: 7, cursor: 'pointer', fontSize: '0.8rem' }}>
          ↻ Refresh
        </button>
      </div>

      {/* KPI Row */}
      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 20 }}>
        <StatCard label="Total Reports"    value={reports.length}             color="var(--accent-cyan)" />
        <StatCard label="Active Analysts"  value={Object.keys(analystMap).length} color="var(--accent-magenta)" />
        <StatCard label="BUY Signals"      value={buys}                       color="#4ade80" />
        <StatCard label="SELL Signals"     value={sells}                      color="#fb923c" />
        <StatCard label="HOLD Signals"     value={holds}                      color="#aaa" />
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 20 }}>
        <Tab label="📋 All Reports" active={tab === 'reports'} onClick={() => setTab('reports')} />
        <Tab label="👤 Users"       active={tab === 'users'}   onClick={() => setTab('users')}   />
        <Tab label="📊 Metrics"     active={tab === 'metrics'} onClick={() => setTab('metrics')} />
      </div>

      {/* ── TAB: REPORTS ──────────────────────────────────────────────────────── */}
      {tab === 'reports' && (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(140px,1fr))', gap: 10, marginBottom: 14 }}>
            <input placeholder="Ticker" value={fTicker} onChange={e => setFTicker(e.target.value)}
              style={{ background:'#111', border:'1px solid #333', color:'#fff', padding:'7px 10px', borderRadius:6, fontSize:'0.85rem' }} />
            <select value={fSignal} onChange={e => setFSignal(e.target.value)}
              style={{ background:'#111', border:'1px solid #333', color:'#fff', padding:'7px 10px', borderRadius:6, fontSize:'0.85rem' }}>
              <option value="">All Signals</option>
              <option value="BUY">BUY</option>
              <option value="SELL">SELL</option>
              <option value="HOLD">HOLD</option>
            </select>
            <input type="date" value={fFrom} onChange={e => setFFrom(e.target.value)}
              style={{ background:'#111', border:'1px solid #333', color:'#fff', padding:'7px 10px', borderRadius:6, fontSize:'0.85rem' }} />
            <input type="date" value={fTo} onChange={e => setFTo(e.target.value)}
              style={{ background:'#111', border:'1px solid #333', color:'#fff', padding:'7px 10px', borderRadius:6, fontSize:'0.85rem' }} />
            <div style={{ display:'flex', gap:8 }}>
              <button onClick={() => { setFTicker(''); setFSignal(''); setFFrom(''); setFTo(''); }}
                style={{ flex:1, background:'transparent', border:'1px solid #444', color:'#aaa', borderRadius:6, cursor:'pointer', fontSize:'0.82rem' }}>Clear</button>
              <button onClick={exportCSV} disabled={filtered.length === 0}
                style={{ flex:1, background:'#1b4332', border:'none', color:'#4ade80', borderRadius:6, cursor:'pointer', fontWeight:700, fontSize:'0.82rem' }}>Export CSV</button>
            </div>
          </div>

          <p style={{ fontSize:'0.8rem', color:'var(--text-muted)', marginBottom:10 }}>
            Showing {filtered.length} of {reports.length} total records
          </p>

          <div style={{ overflowX:'auto' }}>
            <table style={{ width:'100%', borderCollapse:'collapse', textAlign:'left', fontSize:'0.85rem' }}>
              <thead>
                <tr style={{ borderBottom:'2px solid #333', color:'var(--accent-magenta)' }}>
                  {['Date', 'Analyst', 'Ticker', 'Signal', 'PDF'].map(h =>
                    <th key={h} style={{ padding:'10px 12px' }}>{h}</th>
                  )}
                </tr>
              </thead>
              <tbody>
                {filtered.map((r, i) => {
                  const sc = SIGNAL_COLOR(r.signal);
                  return (
                    <tr key={i} className="table-row-hover" style={{ borderBottom:'1px solid #1a1a1a' }}>
                      <td style={{ padding:'10px 12px', color:'var(--text-muted)' }}>{new Date(r.ts).toLocaleString()}</td>
                      <td style={{ padding:'10px 12px', color:'var(--accent-cyan)', fontSize:'0.8rem' }}>{r.analyst}</td>
                      <td style={{ padding:'10px 12px', fontWeight:700 }}>{r.ticker}</td>
                      <td style={{ padding:'10px 12px' }}>
                        <span style={{ padding:'3px 8px', borderRadius:4, background:sc.bg, color:sc.text, fontWeight:700, fontSize:'0.78rem' }}>{r.signal}</span>
                      </td>
                      <td style={{ padding:'10px 12px' }}>
                        <a href={`${API}${r.pdf_file}`} target="_blank" rel="noreferrer"
                          style={{ color:'var(--accent-cyan)', textDecoration:'none', fontWeight:600 }}>View ↗</a>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {filtered.length === 0 && <p style={{ textAlign:'center', color:'var(--text-muted)', marginTop:30 }}>No records match the current filters.</p>}
          </div>
        </>
      )}

      {/* ── TAB: USERS ────────────────────────────────────────────────────────── */}
      {tab === 'users' && (
        <div>
          <p style={{ fontSize:'0.8rem', color:'var(--text-muted)', marginBottom:12 }}>
            Click any analyst to view their individual metrics.
          </p>
          <div style={{ overflowX:'auto' }}>
            <table style={{ width:'100%', borderCollapse:'collapse', textAlign:'left', fontSize:'0.85rem' }}>
              <thead>
                <tr style={{ borderBottom:'2px solid #333', color:'var(--accent-magenta)' }}>
                  {['Analyst', 'Reports', 'BUY', 'SELL', 'HOLD', 'Last Active', ''].map(h =>
                    <th key={h} style={{ padding:'10px 12px' }}>{h}</th>
                  )}
                </tr>
              </thead>
              <tbody>
                {Object.entries(analystMap)
                  .sort((a,b) => b[1].total - a[1].total)
                  .map(([email, stats], i) => {
                    const lastReport = reports.filter(r => r.analyst === email)[0];
                    const isSelected = selectedUser === email;
                    return (
                      <tr key={i} className="table-row-hover"
                        style={{ borderBottom:'1px solid #1a1a1a', cursor:'pointer',
                          background: isSelected ? '#1a1a2e' : 'transparent' }}
                        onClick={() => setSelectedUser(isSelected ? null : email)}>
                        <td style={{ padding:'10px 12px', color:'var(--accent-cyan)', fontWeight:600 }}>{email}</td>
                        <td style={{ padding:'10px 12px', fontWeight:700 }}>{stats.total}</td>
                        <td style={{ padding:'10px 12px', color:'#4ade80' }}>{stats.buy}</td>
                        <td style={{ padding:'10px 12px', color:'#fb923c' }}>{stats.sell}</td>
                        <td style={{ padding:'10px 12px', color:'#aaa' }}>{stats.hold}</td>
                        <td style={{ padding:'10px 12px', color:'var(--text-muted)', fontSize:'0.8rem' }}>
                          {lastReport ? new Date(lastReport.ts).toLocaleDateString() : '—'}
                        </td>
                        <td style={{ padding:'10px 12px', color:'var(--accent-magenta)', fontSize:'0.8rem' }}>
                          {isSelected ? '▲ Hide' : '▼ Details'}
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
            </table>
          </div>

          {/* User Drill-Down */}
          {selectedUser && (() => {
            const userReports = reports.filter(r => r.analyst === selectedUser);
            const uBuys  = userReports.filter(r => r.signal?.includes('BUY')).length;
            const uSells = userReports.filter(r => r.signal?.includes('SELL')).length;
            const uHolds = userReports.filter(r => r.signal?.includes('HOLD')).length;
            const uTickerMap = {};
            userReports.forEach(r => { uTickerMap[r.ticker] = (uTickerMap[r.ticker] || 0) + 1; });
            const uTopTickers = Object.entries(uTickerMap).sort((a,b) => b[1]-a[1]).slice(0,8);
            const uMax = uTopTickers[0]?.[1] || 1;
            return (
              <div style={{ marginTop:20, padding:20, background:'#0d0d0d', border:'1px solid var(--accent-magenta)', borderRadius:10 }}>
                <h3 style={{ color:'var(--accent-magenta)', margin:'0 0 6px' }}>Metrics for: {selectedUser}</h3>
                <p style={{ color:'var(--text-muted)', fontSize:'0.8rem', marginTop:0, marginBottom:20 }}>
                  {userReports.length} total reports
                </p>
                <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:20 }}>
                  <div>
                    <h4 style={{ color:'var(--accent-cyan)', margin:'0 0 12px' }}>Signal Distribution</h4>
                    <BarRow label="BUY"  count={uBuys}  max={userReports.length} color="#4ade80" />
                    <BarRow label="SELL" count={uSells} max={userReports.length} color="#fb923c" />
                    <BarRow label="HOLD" count={uHolds} max={userReports.length} color="#6b7280" />
                    <p style={{ fontSize:'0.75rem', color:'var(--text-muted)', marginTop:12 }}>
                      Bullish rate: <strong style={{ color:'#4ade80' }}>
                        {userReports.length > 0 ? ((uBuys / userReports.length)*100).toFixed(1) : 0}%
                      </strong>
                    </p>
                  </div>
                  <div>
                    <h4 style={{ color:'var(--accent-cyan)', margin:'0 0 12px' }}>Top Tickers</h4>
                    {uTopTickers.map(([ticker, count]) => (
                      <BarRow key={ticker} label={ticker} count={count} max={uMax} color="var(--accent-magenta)" />
                    ))}
                  </div>
                </div>
                {/* Their recent reports */}
                <h4 style={{ color:'var(--accent-cyan)', marginTop:20, marginBottom:10 }}>Recent Reports</h4>
                <div style={{ overflowX:'auto' }}>
                  <table style={{ width:'100%', borderCollapse:'collapse', fontSize:'0.82rem' }}>
                    <thead>
                      <tr style={{ borderBottom:'1px solid #333', color:'#666' }}>
                        {['Date','Ticker','Signal','PDF'].map(h => <th key={h} style={{ padding:'8px 10px', textAlign:'left' }}>{h}</th>)}
                      </tr>
                    </thead>
                    <tbody>
                      {userReports.slice(0,10).map((r, i) => {
                        const sc = SIGNAL_COLOR(r.signal);
                        return (
                          <tr key={i} style={{ borderBottom:'1px solid #1a1a1a' }}>
                            <td style={{ padding:'8px 10px', color:'var(--text-muted)' }}>{new Date(r.ts).toLocaleDateString()}</td>
                            <td style={{ padding:'8px 10px', fontWeight:700 }}>{r.ticker}</td>
                            <td style={{ padding:'8px 10px' }}>
                              <span style={{ padding:'2px 7px', borderRadius:4, background:sc.bg, color:sc.text, fontSize:'0.75rem', fontWeight:700 }}>{r.signal}</span>
                            </td>
                            <td style={{ padding:'8px 10px' }}>
                              <a href={`${API}${r.pdf_file}`} target="_blank" rel="noreferrer"
                                style={{ color:'var(--accent-cyan)', textDecoration:'none' }}>View ↗</a>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            );
          })()}
        </div>
      )}

      {/* ── TAB: METRICS ──────────────────────────────────────────────────────── */}
      {tab === 'metrics' && (
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:20 }}>

          {/* Signal distribution */}
          <div style={{ padding:20, background:'#0d0d0d', border:'1px solid #2a2a2a', borderRadius:10 }}>
            <h3 style={{ color:'var(--accent-cyan)', margin:'0 0 16px' }}>Signal Distribution</h3>
            <BarRow label="BUY"  count={buys}  max={reports.length} color="#4ade80" />
            <BarRow label="SELL" count={sells} max={reports.length} color="#fb923c" />
            <BarRow label="HOLD" count={holds} max={reports.length} color="#6b7280" />
            <p style={{ fontSize:'0.75rem', color:'var(--text-muted)', marginTop:12 }}>
              Bullish rate: <strong style={{ color:'#4ade80' }}>
                {reports.length > 0 ? ((buys / reports.length) * 100).toFixed(1) : 0}%
              </strong>
            </p>
          </div>

          {/* Top tickers */}
          <div style={{ padding:20, background:'#0d0d0d', border:'1px solid #2a2a2a', borderRadius:10 }}>
            <h3 style={{ color:'var(--accent-cyan)', margin:'0 0 16px' }}>Top 10 Most Analyzed</h3>
            {topTickers.map(([ticker, count]) => (
              <BarRow key={ticker} label={ticker} count={count} max={maxTicker} color="var(--accent-magenta)" />
            ))}
            {topTickers.length === 0 && <p style={{ color:'var(--text-muted)' }}>No data yet.</p>}
          </div>

          {/* Analyst output summary */}
          <div style={{ padding:20, background:'#0d0d0d', border:'1px solid #2a2a2a', borderRadius:10, gridColumn:'span 2' }}>
            <h3 style={{ color:'var(--accent-cyan)', margin:'0 0 16px' }}>Analyst Output Ranking</h3>
            {Object.entries(analystMap)
              .sort((a,b) => b[1].total - a[1].total)
              .map(([email, s]) => (
                <BarRow key={email} label={email} count={s.total} max={Math.max(...Object.values(analystMap).map(v => v.total))} color="var(--accent-cyan)" />
              ))}
          </div>

        </div>
      )}
    </div>
  );
}

export default Admin;
