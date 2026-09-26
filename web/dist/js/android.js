function switchView(viewId, btn) {
  // Hide all views
  document.querySelectorAll('.view-section').forEach(el => el.style.display = 'none');
  // Show selected view
  document.getElementById('view-' + viewId).style.display = 'block';

  // Update nav buttons
  document.querySelectorAll('.nav-btn').forEach(el => el.classList.remove('active'));
  btn.classList.add('active');

  if (viewId === 'received') {
    fetchHistory();
  }
}

async function fetchHistory() {
  try {
    const res = await fetch('/api/history');
    const files = await res.json();
    const list = document.getElementById('historyList');
    list.innerHTML = '';
    if (!files || files.length === 0) {
      list.innerHTML = '<p>No files received yet.</p>';
      return;
    }
    files.forEach(f => {
      const p = document.createElement('p');
      const sizeMB = (f.size / (1024*1024)).toFixed(2);
      p.innerHTML = `<strong>${f.name}</strong><br><span class="small-hint">${sizeMB} MB • ${new Date(f.time).toLocaleString()}</span>`;
      list.appendChild(p);
    });
  } catch (e) {
    console.error("Error fetching history", e);
  }
}

async function fetchSettings() {
  try {
    const res = await fetch('/api/settings');
    const s = await res.json();
    document.getElementById('setSaveDir').value = s.saveDirectory || '';
    document.getElementById('setMulticastAddr').value = s.multicastAddr || '';
    document.getElementById('setMulticastPort').value = s.multicastPort || '';
    document.getElementById('setDeviceName').value = s.deviceName || '';
    document.getElementById('setBlockedPorts').value = s.blockedPortsIn ? s.blockedPortsIn.join(', ') : '';
  } catch (e) {
    console.error("Error fetching settings", e);
  }
}

async function saveSettings() {
  const saveDir = document.getElementById('setSaveDir').value;
  const mAddr = document.getElementById('setMulticastAddr').value;
  const mPort = parseInt(document.getElementById('setMulticastPort').value);
  const dName = document.getElementById('setDeviceName').value;
  const bPortsStr = document.getElementById('setBlockedPorts').value;
  const bPorts = bPortsStr.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n));

  const settings = {
    saveDirectory: saveDir,
    multicastAddr: mAddr,
    multicastPort: mPort,
    deviceName: dName,
    blockedPortsIn: bPorts,
    blockedPortsOut: [],
    blockedDevices: []
  };

  await fetch('/api/settings', {
    method: 'POST',
    body: JSON.stringify(settings)
  });
  alert('Settings saved!');
}

function refreshSetting(settingName) {
  fetchSettings(); // Simple implementation: refetch all
  alert('Refreshed ' + settingName);
}

function setDefaultSettings() {
  document.getElementById('setSaveDir').value = '/sdcard/Download/FileShare';
  document.getElementById('setMulticastAddr').value = '224.0.0.1';
  document.getElementById('setMulticastPort').value = '53535';
  document.getElementById('setDeviceName').value = 'Android Device';
  document.getElementById('setBlockedPorts').value = '';
}

async function restartServer() {
  // In a real implementation this would call an API endpoint to restart the listener
  alert('Server restarting... You may need to refresh the app.');
}

function showQrModal() {
  alert('QR Code generation would appear here showing the local IP.');
}

// Device Info Populator
window.onload = function() {
  document.getElementById('infoModel').innerText = `Model: ${navigator.userAgent.split(';')[1] || navigator.platform}`;
  document.getElementById('infoOs').innerText = `OS Name: Android`;
  document.getElementById('infoType').innerText = `Type: SMART Phone`;
  
  fetchSettings();
  fetchHistory();
};
