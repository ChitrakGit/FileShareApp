// FileShare Web App Controller
let selectedItems = []; // Array of { file: File, relativePath: string }
let discoveredPeers = [];
let localInfo = { deviceName: 'Loading...', ip: '127.0.0.1', port: 8990 };
let currentUploadXHR = null;
let activeIncomingProposal = null;

// DOM Elements
const dropzone = document.getElementById('dropzone');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const selectionCard = document.getElementById('selectionCard');
const selectionList = document.getElementById('selectionList');
const payloadTitle = document.getElementById('payloadTitle');
const payloadSize = document.getElementById('payloadSize');
const btnClearSelection = document.getElementById('btnClearSelection');
const targetPeerSelect = document.getElementById('targetPeerSelect');
const btnQuickAddDevice = document.getElementById('btnQuickAddDevice');
const quickAddModal = document.getElementById('quickAddModal');
const btnCloseQuickAddModal = document.getElementById('btnCloseQuickAddModal');
const btnCancelQuickAdd = document.getElementById('btnCancelQuickAdd');
const btnConfirmQuickAdd = document.getElementById('btnConfirmQuickAdd');
const quickAddIp = document.getElementById('quickAddIp');
const quickAddName = document.getElementById('quickAddName');
const quickAddStatus = document.getElementById('quickAddStatus');

const optionalPin = document.getElementById('optionalPin');
const btnSendPayload = document.getElementById('btnSendPayload');

const transferProgressCard = document.getElementById('transferProgressCard');
const transferDirectionBadge = document.getElementById('transferDirectionBadge');
const transferItemName = document.getElementById('transferItemName');
const progressBarFill = document.getElementById('progressBarFill');
const transferPercentage = document.getElementById('transferPercentage');
const transferSpeed = document.getElementById('transferSpeed');
const transferTransferred = document.getElementById('transferTransferred');
const transferEta = document.getElementById('transferEta');
const btnCancelTransfer = document.getElementById('btnCancelTransfer');

const peersList = document.getElementById('peersList');
const btnRefreshPeers = document.getElementById('btnRefreshPeers');
const manualPeerIp = document.getElementById('manualPeerIp');
const manualPeerName = document.getElementById('manualPeerName');
const btnAddManualPeer = document.getElementById('btnAddManualPeer');
const scanSubnetInput = document.getElementById('scanSubnetInput');
const btnScanSubnet = document.getElementById('btnScanSubnet');
const peerActionStatus = document.getElementById('peerActionStatus');

const localDeviceName = document.getElementById('localDeviceName');
const localDeviceIp = document.getElementById('localDeviceIp');
const btnCopyUrl = document.getElementById('btnCopyUrl');
const btnQrCode = document.getElementById('btnQrCode');
const qrModal = document.getElementById('qrModal');
const btnCloseQrModal = document.getElementById('btnCloseQrModal');
const qrImage = document.getElementById('qrImage');
const qrUrlText = document.getElementById('qrUrlText');

const btnScanQrCode = document.getElementById('btnScanQrCode');
const scannerModal = document.getElementById('scannerModal');
const btnCloseScannerModal = document.getElementById('btnCloseScannerModal');
let html5QrcodeScanner = null;

const incomingModal = document.getElementById('incomingModal');
const incomingSenderName = document.getElementById('incomingSenderName');
const incomingSenderIp = document.getElementById('incomingSenderIp');
const incomingFileCount = document.getElementById('incomingFileCount');
const incomingTotalSize = document.getElementById('incomingTotalSize');
const incomingItemName = document.getElementById('incomingItemName');
const pinVerificationBox = document.getElementById('pinVerificationBox');
const incomingPinInput = document.getElementById('incomingPinInput');
const btnAcceptTransfer = document.getElementById('btnAcceptTransfer');
const btnRejectTransfer = document.getElementById('btnRejectTransfer');
const historyList = document.getElementById('historyList');

// 1. Initialize and Fetch Local Node Information
async function init() {
  try {
    const res = await fetch('/api/info');
    if (res.ok) {
      localInfo = await res.json();
      localDeviceName.textContent = localInfo.deviceName;
      localDeviceIp.textContent = `http://${localInfo.ip}:${localInfo.port}`;
    }
  } catch (err) {
    console.warn('Running in standalone or server starting:', err);
    localDeviceName.textContent = 'This Device';
    localDeviceIp.textContent = `http://${window.location.host}`;
  }

  fetchPeers();
  setInterval(fetchPeers, 3000);
  setupWebSocket();
  loadHistory();
  
  // Wails Desktop Integrations
  if (window.runtime) {
    window.runtime.EventsOn("mode_changed", (mode) => {
      const appContainer = document.querySelector('.app-container');
      const drawerContainer = document.getElementById('drawerContainer');
      if (mode === "drawer") {
        appContainer.style.display = 'none';
        document.body.style.background = 'transparent';
        document.querySelector('.app-background').style.display = 'none';
        drawerContainer.style.display = 'flex';
        renderDrawerPeers();
      } else {
        appContainer.style.display = 'flex';
        document.body.style.background = 'var(--bg-dark)';
        document.querySelector('.app-background').style.display = 'block';
        drawerContainer.style.display = 'none';
      }
    });
    
    // Bind restore UI button
    document.getElementById('btnRestoreUi')?.addEventListener('click', () => {
      window.runtime.EventsEmit("restore_ui");
    });
  }
}

// 2. Fetch Discovered Peers
async function fetchPeers() {
  try {
    const res = await fetch('/api/peers');
    if (res.ok) {
      const data = await res.json();
      discoveredPeers = data || [];
      renderPeers();
    }
  } catch (err) {
    // Silent fail if backend starting
  }
}

function renderPeers() {
  const currentVal = targetPeerSelect.value;
  targetPeerSelect.innerHTML = '<option value="">Choose a discovered peer...</option>';

  if (discoveredPeers.length === 0) {
    peersList.innerHTML = `
      <div class="empty-peers">
        <div class="radar-animation">
          <div class="radar-circle"></div>
          <div class="radar-circle circle-2"></div>
        </div>
        <p>Scanning Wi-Fi for devices running FileShare...</p>
      </div>
    `;
    return;
  }

  peersList.innerHTML = '';
  discoveredPeers.forEach(peer => {
    // Add to dropdown
    const opt = document.createElement('option');
    opt.value = `${peer.ip}:${peer.port}`;
    const savedTag = peer.isSaved ? ' [Saved]' : '';
    opt.textContent = `${peer.name} (${peer.ip}:${peer.port}) - ${peer.os || 'LAN'}${savedTag}`;
    if (opt.value === currentVal) opt.selected = true;
    targetPeerSelect.appendChild(opt);

    // Add to radar list
    const card = document.createElement('div');
    card.className = 'peer-card';
    card.innerHTML = `
      <div class="peer-info">
        <div class="peer-avatar">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
        </div>
        <div>
          <div class="peer-name">
            ${escapeHtml(peer.name)}
            ${peer.isSaved ? '<span class="badge-saved">Saved</span>' : ''}
          </div>
          <div class="peer-meta">${escapeHtml(peer.ip)}:${peer.port} • ${escapeHtml(peer.os || 'LAN')}</div>
        </div>
      </div>
      <div class="peer-actions">
        <button class="btn btn-sm btn-secondary select-peer-btn">Select</button>
        ${peer.isSaved ? `
          <button class="btn-delete-peer" title="Remove saved device" data-id="${escapeHtml(peer.id || peer.ip)}">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
          </button>
        ` : ''}
      </div>
    `;

    card.querySelector('.select-peer-btn').addEventListener('click', () => {
      targetPeerSelect.value = `${peer.ip}:${peer.port}`;
      checkSendButtonStatus();
      card.style.borderColor = 'var(--primary)';
      setTimeout(() => card.style.borderColor = '', 1000);
    });

    const deleteBtn = card.querySelector('.btn-delete-peer');
    if (deleteBtn) {
      deleteBtn.addEventListener('click', async (e) => {
        e.stopPropagation();
        await removeSavedPeer(deleteBtn.dataset.id);
      });
    }

    peersList.appendChild(card);
  });

  checkSendButtonStatus();
  
  if (document.getElementById('drawerContainer').style.display === 'flex') {
    renderDrawerPeers();
  }
}

// 2b. Render Grouped Peers for Desktop Drawer
function renderDrawerPeers() {
  const drawerList = document.getElementById('drawerPeersList');
  if (!drawerList) return;
  drawerList.innerHTML = '';
  
  // Group by OS
  const groups = {
    'windows': [],
    'darwin': [],
    'linux': [],
    'mobile': [],
    'tablet': [],
    'other': []
  };
  
  discoveredPeers.forEach(p => {
    let os = (p.os || 'other').toLowerCase();
    if (os.includes('android') || os.includes('ios')) {
       os = 'mobile';
    } else if (os.includes('ipad')) {
       os = 'tablet';
    } else if (!groups[os]) {
       os = 'other';
    }
    groups[os].push(p);
  });
  
  Object.keys(groups).forEach(osType => {
    const peers = groups[osType];
    if (peers.length === 0) return;
    
    // Determine icon based on OS
    let svgIcon = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>'; // Desktop
    if (osType === 'mobile') {
      svgIcon = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"></rect><line x1="12" y1="18" x2="12.01" y2="18"></line></svg>';
    } else if (osType === 'tablet') {
      svgIcon = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="2" width="16" height="20" rx="2" ry="2"></rect><line x1="12" y1="18" x2="12.01" y2="18"></line></svg>';
    }
    
    // The group represents either a single device or multiple
    const isActive = peers.some(p => p.online !== false);
    
    const div = document.createElement('div');
    div.className = 'drawer-peer-item';
    div.innerHTML = `
      <div class="drawer-peer-icon">
        ${svgIcon}
        ${peers.length > 1 ? \`<div class="drawer-badge \${isActive ? 'active' : 'inactive'}">\${peers.length}</div>\` : \`<div class="drawer-badge \${isActive ? 'active' : 'inactive'}" style="width:12px;height:12px;"></div>\`}
      </div>
      <div class="drawer-peer-name">${peers.length === 1 ? escapeHtml(peers[0].name) : peers.length + ' Devices'}</div>
    `;
    
    // Setup drag and drop for this group
    div.addEventListener('dragover', (e) => {
      e.preventDefault();
      div.classList.add('dragover');
    });
    
    div.addEventListener('dragleave', () => {
      div.classList.remove('dragover');
    });
    
    div.addEventListener('drop', (e) => {
      e.preventDefault();
      div.classList.remove('dragover');
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
         // Auto-select the first active peer in this group and send
         const target = peers.find(p => p.online !== false) || peers[0];
         targetPeerSelect.value = \`\${target.ip}:\${target.port}\`;
         // Dispatch files to the main drop handler by faking an event to the main dropzone
         window.runtime.EventsEmit("restore_ui"); // Restore UI to show progress
         setTimeout(() => {
             const dt = new DataTransfer();
             for(let i=0; i<e.dataTransfer.files.length; i++) {
                 dt.items.add(e.dataTransfer.files[i]);
             }
             fileInput.files = dt.files;
             handleFiles(fileInput.files);
             btnSendPayload.click();
         }, 500);
      }
    });
    
    drawerList.appendChild(div);
  });
}

// 3. Manual Peer Addition & Subnet Scanning
btnAddManualPeer.addEventListener('click', async () => {
  let ip = manualPeerIp.value.trim();
  // Strip http://, https://, and any trailing slashes
  ip = ip.replace(/^https?:\/\//i, '').replace(/\/+$/, '');
  if (!ip) {
    showActionStatus('Please enter an IP address or hostname.', 'error');
    return;
  }

  const name = manualPeerName.value.trim();
  showActionStatus('Connecting to device to verify...', 'info');
  btnAddManualPeer.disabled = true;

  try {
    const res = await fetch('/api/peers/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ address: ip, name: name })
    });

    if (res.ok) {
      const saved = await res.json();
      showActionStatus(`Successfully saved device: ${saved.name} (${saved.ip}:${saved.port})!`, 'success');
      manualPeerIp.value = '';
      manualPeerName.value = '';
      await fetchPeers();
      targetPeerSelect.value = `${saved.ip}:${saved.port}`;
      checkSendButtonStatus();
    } else {
      const errText = await res.text();
      showActionStatus(`Could not reach device: ${errText}`, 'error');
    }
  } catch (err) {
    showActionStatus(`Network error: ${err.message}`, 'error');
  } finally {
    btnAddManualPeer.disabled = false;
  }
});

btnScanSubnet.addEventListener('click', async () => {
  const subnet = scanSubnetInput.value.trim();
  showActionStatus(`Scanning subnet ${subnet || 'local'} (1-254)...`, 'info');
  btnScanSubnet.disabled = true;

  try {
    const res = await fetch('/api/peers/scan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ subnet: subnet })
    });

    if (res.ok) {
      const found = await res.json();
      showActionStatus(`Scan complete! Found ${found ? found.length : 0} device(s).`, 'success');
      await fetchPeers();
    } else {
      showActionStatus('Subnet scan failed.', 'error');
    }
  } catch (err) {
    showActionStatus(`Scan error: ${err.message}`, 'error');
  } finally {
    btnScanSubnet.disabled = false;
  }
});

async function removeSavedPeer(id) {
  try {
    const res = await fetch(`/api/peers/remove?id=${encodeURIComponent(id)}`, {
      method: 'POST'
    });
    if (res.ok) {
      showActionStatus('Saved device removed.', 'info');
      await fetchPeers();
    }
  } catch (err) {
    console.error('Failed to remove saved peer:', err);
  }
}

function showActionStatus(msg, type) {
  peerActionStatus.style.display = 'block';
  peerActionStatus.textContent = msg;
  if (type === 'error') {
    peerActionStatus.style.color = '#f87171';
  } else if (type === 'success') {
    peerActionStatus.style.color = '#34d399';
  } else {
    peerActionStatus.style.color = '#94a3b8';
  }
}

// 4. Drag & Drop and File/Folder Inputs
dropzone.addEventListener('dragover', (e) => {
  e.preventDefault();
  dropzone.classList.add('dragover');
});

dropzone.addEventListener('dragleave', () => {
  dropzone.classList.remove('dragover');
});

dropzone.addEventListener('drop', async (e) => {
  e.preventDefault();
  dropzone.classList.remove('dragover');

  const items = e.dataTransfer.items;
  if (items && items.length > 0) {
    const collectedFiles = [];
    const queue = [];

    for (let i = 0; i < items.length; i++) {
      const entry = items[i].webkitGetAsEntry ? items[i].webkitGetAsEntry() : null;
      if (entry) {
        queue.push(traverseFileTree(entry, ''));
      } else {
        const file = items[i].getAsFile();
        if (file) collectedFiles.push({ file, relativePath: file.name });
      }
    }

    const results = await Promise.all(queue);
    results.forEach(resList => collectedFiles.push(...resList));
    addFilesToSelection(collectedFiles);
  } else if (e.dataTransfer.files) {
    const list = Array.from(e.dataTransfer.files).map(f => ({ file: f, relativePath: f.name }));
    addFilesToSelection(list);
  }
});

async function traverseFileTree(item, path) {
  return new Promise((resolve) => {
    if (item.isFile) {
      item.file((file) => {
        resolve([{ file, relativePath: path + file.name }]);
      }, () => resolve([]));
    } else if (item.isDirectory) {
      const dirReader = item.createReader();
      const entries = [];

      const readEntries = () => {
        dirReader.readEntries(async (result) => {
          if (!result.length) {
            const subPromises = entries.map(subItem => traverseFileTree(subItem, path + item.name + '/'));
            const subResults = await Promise.all(subPromises);
            resolve(subResults.flat());
          } else {
            entries.push(...result);
            readEntries();
          }
        }, () => resolve([]));
      };
      readEntries();
    } else {
      resolve([]);
    }
  });
}

fileInput.addEventListener('change', () => {
  if (!fileInput.files.length) return;
  const list = Array.from(fileInput.files).map(f => ({ file: f, relativePath: f.name }));
  addFilesToSelection(list);
  fileInput.value = '';
});

folderInput.addEventListener('change', () => {
  if (!folderInput.files.length) return;
  const list = Array.from(folderInput.files).map(f => ({
    file: f,
    relativePath: f.webkitRelativePath || f.name
  }));
  addFilesToSelection(list);
  folderInput.value = '';
});

function addFilesToSelection(newItems) {
  selectedItems.push(...newItems);
  updateSelectionUI();
}

function updateSelectionUI() {
  if (selectedItems.length === 0) {
    selectionCard.style.display = 'none';
    btnSendPayload.disabled = true;
    return;
  }

  selectionCard.style.display = 'block';
  let totalBytes = 0;
  selectionList.innerHTML = '';

  selectedItems.forEach((item, idx) => {
    totalBytes += item.file.size;
    const row = document.createElement('div');
    row.className = 'item-row';
    row.innerHTML = `
      <div class="item-name-group">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path></svg>
        <span title="${escapeHtml(item.relativePath)}">${escapeHtml(item.relativePath)}</span>
      </div>
      <span class="item-size">${formatBytes(item.file.size)}</span>
    `;
    selectionList.appendChild(row);
  });

  payloadTitle.textContent = `${selectedItems.length} item${selectedItems.length > 1 ? 's' : ''} ready to send`;
  payloadSize.textContent = `Total: ${formatBytes(totalBytes)}`;
  checkSendButtonStatus();
}

btnClearSelection.addEventListener('click', () => {
  selectedItems = [];
  updateSelectionUI();
});

targetPeerSelect.addEventListener('change', checkSendButtonStatus);

function checkSendButtonStatus() {
  btnSendPayload.disabled = !(selectedItems.length > 0 && targetPeerSelect.value);
}

// 4b. Quick Add Device Modal Logic
if (btnQuickAddDevice) {
  btnQuickAddDevice.addEventListener('click', () => {
    quickAddModal.style.display = 'flex';
    quickAddIp.value = '';
    quickAddName.value = '';
    quickAddStatus.style.display = 'none';
    quickAddIp.focus();
  });
}

function closeQuickAddModal() {
  quickAddModal.style.display = 'none';
}

if (btnCloseQuickAddModal) btnCloseQuickAddModal.addEventListener('click', closeQuickAddModal);
if (btnCancelQuickAdd) btnCancelQuickAdd.addEventListener('click', closeQuickAddModal);

if (btnConfirmQuickAdd) {
  btnConfirmQuickAdd.addEventListener('click', async () => {
    let ip = quickAddIp.value.trim();
    ip = ip.replace(/^https?:\/\//i, '').replace(/\/+$/, '');
    if (!ip) {
      quickAddStatus.style.display = 'block';
      quickAddStatus.style.color = '#f87171';
      quickAddStatus.textContent = 'Please enter an IP address or URL.';
      return;
    }

    const name = quickAddName.value.trim();
    quickAddStatus.style.display = 'block';
    quickAddStatus.style.color = '#3b82f6';
    quickAddStatus.textContent = 'Verifying device connection...';
    btnConfirmQuickAdd.disabled = true;

    try {
      const res = await fetch('/api/peers/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ address: ip, name: name })
      });

      if (res.ok) {
        const saved = await res.json();
        await fetchPeers();
        targetPeerSelect.value = `${saved.ip}:${saved.port}`;
        checkSendButtonStatus();
        closeQuickAddModal();
        showActionStatus(`Successfully saved device: ${saved.name}!`, 'success');
      } else {
        const errText = await res.text();
        quickAddStatus.style.color = '#f87171';
        quickAddStatus.textContent = `Could not reach device: ${errText}`;
      }
    } catch (err) {
      quickAddStatus.style.color = '#f87171';
      quickAddStatus.textContent = `Network error: ${err.message}`;
    } finally {
      btnConfirmQuickAdd.disabled = false;
    }
  });
}

// 5. Send Payload to Remote Peer
btnSendPayload.addEventListener('click', async () => {
  if (selectedItems.length === 0 || !targetPeerSelect.value) return;

  const target = targetPeerSelect.value;
  const pin = optionalPin.value.trim();

  // Prepare FormData
  const formData = new FormData();
  let totalBytes = 0;

  selectedItems.forEach(item => {
    totalBytes += item.file.size;
    formData.append('files', item.file);
    formData.append('paths', item.relativePath);
  });

  if (pin) formData.append('pin', pin);

  // Show live progress UI
  showProgressUI('Sending', `${selectedItems.length} items to ${target}`, totalBytes);

  const targetUrl = target.startsWith('http') ? target : `http://${target}`;
  const uploadUrl = `${targetUrl}/api/upload`;

  const startTime = Date.now();
  let lastLoaded = 0;
  let lastTime = startTime;

  currentUploadXHR = new XMLHttpRequest();
  currentUploadXHR.open('POST', uploadUrl, true);

  currentUploadXHR.upload.onprogress = (e) => {
    if (e.lengthComputable) {
      const percent = Math.round((e.loaded / e.total) * 100);
      progressBarFill.style.width = `${percent}%`;
      transferPercentage.textContent = `${percent}%`;
      transferTransferred.textContent = `${formatBytes(e.loaded)} / ${formatBytes(e.total)}`;

      // Calculate speed and ETA
      const now = Date.now();
      const elapsedSec = (now - lastTime) / 1000;
      if (elapsedSec >= 0.5) {
        const bytesDiff = e.loaded - lastLoaded;
        const speedBps = bytesDiff / elapsedSec;
        transferSpeed.textContent = `${formatBytes(speedBps)}/s`;

        const remainingBytes = e.total - e.loaded;
        const etaSec = speedBps > 0 ? Math.ceil(remainingBytes / speedBps) : 0;
        transferEta.textContent = `ETA: ${etaSec}s`;

        lastLoaded = e.loaded;
        lastTime = now;
      }
    }
  };

  currentUploadXHR.onload = () => {
    if (currentUploadXHR.status >= 200 && currentUploadXHR.status < 300) {
      progressBarFill.style.width = '100%';
      transferPercentage.textContent = '100%';
      transferSpeed.textContent = 'Completed!';
      transferEta.textContent = 'Success';

      addHistoryRecord({
        name: `${selectedItems.length} items sent`,
        size: totalBytes,
        peer: target,
        type: 'Sent',
        time: new Date().toLocaleTimeString()
      });

      setTimeout(() => {
        hideProgressUI();
        selectedItems = [];
        updateSelectionUI();
      }, 1500);
    } else {
      alert(`Transfer failed: ${currentUploadXHR.responseText || 'Check network connection or PIN'}`);
      hideProgressUI();
    }
  };

  currentUploadXHR.onerror = () => {
    alert('Network error during transfer. Ensure the recipient is reachable.');
    hideProgressUI();
  };

  currentUploadXHR.send(formData);
});

btnCancelTransfer.addEventListener('click', () => {
  if (currentUploadXHR) {
    currentUploadXHR.abort();
    currentUploadXHR = null;
    hideProgressUI();
  }
});

function showProgressUI(direction, title, totalBytes) {
  transferProgressCard.style.display = 'block';
  transferDirectionBadge.textContent = direction;
  transferItemName.textContent = title;
  progressBarFill.style.width = '0%';
  transferPercentage.textContent = '0%';
  transferTransferred.textContent = `0 B / ${formatBytes(totalBytes)}`;
  transferSpeed.textContent = 'Starting...';
  transferEta.textContent = 'Calculating...';
}

function hideProgressUI() {
  transferProgressCard.style.display = 'none';
}

// 6. WebSocket for Live Signals and Incoming Push Notifications
function setupWebSocket() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws`;

  let ws;
  try {
    ws = new WebSocket(wsUrl);
  } catch (e) {
    return;
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === 'peer_update') {
        discoveredPeers = msg.peers || [];
        renderPeers();
      } else if (msg.type === 'transfer_progress') {
        // Handle server-side download progress
        transferProgressCard.style.display = 'block';
        transferDirectionBadge.textContent = 'Receiving';
        transferItemName.textContent = 'Incoming Transfer...';
        
        const percent = Math.round(msg.percent);
        progressBarFill.style.width = `${percent}%`;
        transferPercentage.textContent = `${percent}%`;
        transferSpeed.textContent = `${msg.speedMBps.toFixed(1)} MB/s`;
        transferEta.textContent = `ETA: ${msg.etaSec}s`;
        
        if (percent >= 100) {
           setTimeout(() => {
             hideProgressUI();
           }, 2000);
        }
      }
    } catch (err) {}
  };

  ws.onclose = () => {
    setTimeout(setupWebSocket, 4000);
  };
}

// 7. QR Code Modal for Mobile Devices
btnQrCode.addEventListener('click', () => {
  const localUrl = `http://${localInfo.ip || window.location.hostname}:${localInfo.port || 8990}`;
  qrUrlText.textContent = localUrl;
  // Use local or dynamic QR generator
  qrImage.src = `/api/qr?url=${encodeURIComponent(localUrl)}`;
  qrModal.style.display = 'flex';
});

btnCloseQrModal.addEventListener('click', () => {
  qrModal.style.display = 'none';
});

// 7b. QR Code Scanner for Peer Connection
if (btnScanQrCode) {
  btnScanQrCode.addEventListener('click', () => {
    scannerModal.style.display = 'flex';
    
    if (!html5QrcodeScanner) {
      html5QrcodeScanner = new Html5QrcodeScanner(
        "reader", { fps: 10, qrbox: 250 }, false);
    }
    
    html5QrcodeScanner.render(onScanSuccess, onScanFailure);
  });
}

if (btnCloseScannerModal) {
  btnCloseScannerModal.addEventListener('click', () => {
    scannerModal.style.display = 'none';
    if (html5QrcodeScanner) {
      html5QrcodeScanner.clear().catch(error => {
        console.error("Failed to clear scanner", error);
      });
    }
  });
}

function onScanSuccess(decodedText, decodedResult) {
  // decodedText should be the URL like http://192.168.1.15:8990
  if (html5QrcodeScanner) {
    html5QrcodeScanner.clear();
  }
  scannerModal.style.display = 'none';
  
  // Clean up URL and auto-add
  let ip = decodedText.trim().replace(/^https?:\/\//i, '').replace(/\/+$/, '');
  manualPeerIp.value = ip;
  showActionStatus('QR Code scanned successfully! Connecting...', 'success');
  
  // Trigger the add button directly
  btnAddManualPeer.click();
}

function onScanFailure(error) {
  // handle scan failure, usually better to ignore and keep scanning
}

if (btnCopyUrl) {
  btnCopyUrl.addEventListener('click', () => {
    const url = localDeviceIp.textContent;
    navigator.clipboard.writeText(url).then(() => {
      const originalTitle = btnCopyUrl.title;
      btnCopyUrl.title = "Copied!";
      setTimeout(() => btnCopyUrl.title = originalTitle, 2000);
    });
  });
}

const btnConfigureFirewall = document.getElementById('btnConfigureFirewall');
if (btnConfigureFirewall) {
  btnConfigureFirewall.addEventListener('click', async () => {
    btnConfigureFirewall.disabled = true;
    showActionStatus('Requesting Administrator permissions...', 'info');
    try {
      const res = await fetch('/api/firewall', { method: 'POST' });
      if (res.ok) {
        showActionStatus('Firewall configured successfully!', 'success');
      } else {
        showActionStatus('Failed to configure firewall.', 'error');
      }
    } catch (err) {
      showActionStatus(`Firewall error: ${err.message}`, 'error');
    } finally {
      btnConfigureFirewall.disabled = false;
    }
  });
}

// 8. History Storage
function loadHistory() {
  const records = JSON.parse(localStorage.getItem('fileshare_history') || '[]');
  if (records.length === 0) {
    historyList.innerHTML = '<div class="empty-history"><p>No transfers yet in this session</p></div>';
    return;
  }

  historyList.innerHTML = '';
  records.slice(0, 10).forEach(r => {
    const row = document.createElement('div');
    row.className = 'item-row';
    row.innerHTML = `
      <div class="item-name-group">
        <span class="badge ${r.type === 'Sent' ? '' : 'badge-tag'}">${r.type}</span>
        <span>${escapeHtml(r.name)} (${formatBytes(r.size)})</span>
      </div>
      <span class="item-size">${escapeHtml(r.time)}</span>
    `;
    historyList.appendChild(row);
  });
}

function addHistoryRecord(record) {
  const records = JSON.parse(localStorage.getItem('fileshare_history') || '[]');
  records.unshift(record);
  localStorage.setItem('fileshare_history', JSON.stringify(records.slice(0, 30)));
  loadHistory();
}

// Helpers
function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/[&<>'"]/g, tag => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    "'": '&#39;',
    '"': '&quot;'
  }[tag] || tag));
}

// Start
window.addEventListener('DOMContentLoaded', init);

// PC Settings Logic
async function fetchSettingsPC() {
  try {
    const res = await fetch('/api/settings');
    const s = await res.json();
    document.getElementById('pcSetSaveDir').value = s.saveDirectory || '';
    document.getElementById('pcSetMulticastAddr').value = s.multicastAddr || '';
    document.getElementById('pcSetMulticastPort').value = s.multicastPort || '';
    document.getElementById('pcSetDeviceName').value = s.deviceName || '';
    document.getElementById('pcSetBlockedPorts').value = s.blockedPortsIn ? s.blockedPortsIn.join(', ') : '';

    // Set OS and Model Info
    document.getElementById('pcInfoOs').innerText = `OS Name: ${navigator.platform}`;
  } catch (e) {
    console.error("Error fetching settings", e);
  }
}

async function saveSettingsPC() {
  const saveDir = document.getElementById('pcSetSaveDir').value;
  const mAddr = document.getElementById('pcSetMulticastAddr').value;
  const mPort = parseInt(document.getElementById('pcSetMulticastPort').value);
  const dName = document.getElementById('pcSetDeviceName').value;
  const bPortsStr = document.getElementById('pcSetBlockedPorts').value;
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

  try {
    await fetch('/api/settings', {
      method: 'POST',
      body: JSON.stringify(settings)
    });
    alert('Settings saved!');
  } catch(e) {
    alert('Failed to save settings');
  }
}

function setDefaultSettingsPC() {
  document.getElementById('pcSetSaveDir').value = 'C:\\Downloads\\FileShare';
  document.getElementById('pcSetMulticastAddr').value = '224.0.0.1';
  document.getElementById('pcSetMulticastPort').value = '53535';
  document.getElementById('pcSetDeviceName').value = 'Desktop PC';
  document.getElementById('pcSetBlockedPorts').value = '';
}


// --- PC App View Switching ---
function switchViewPC(viewId, btnElement) {
    // Hide all views
    document.querySelectorAll('.app-view').forEach(el => el.style.display = 'none');
    document.querySelectorAll('.app-view').forEach(el => el.classList.remove('active'));

    // Remove active class from all tabs
    document.querySelectorAll('.pc-tab-btn').forEach(el => el.classList.remove('active'));

    // Show target view
    const targetView = document.getElementById('view-' + viewId);
    if (targetView) {
        targetView.style.display = 'block';
        targetView.classList.add('active');
    }

    // Set active tab
    if (btnElement) {
        btnElement.classList.add('active');
    }

    // Special handlers based on view
    if (viewId === 'settings') {
        fetchSettingsPC();
    }
}

// --- Wails Drawer Expand ---
function expandDrawerApp() {
    console.log("Expanding Drawer App");
    
    // Hide the drawer
    const drawer = document.getElementById('drawerContainer');
    if (drawer) {
        drawer.style.display = 'none';
    }

    // Try to resize the Wails window if running in Desktop mode
    if (window.runtime && window.runtime.WindowSetSize) {
        try {
            window.runtime.WindowSetSize(1024, 768);
            window.runtime.WindowCenter();
        } catch (e) {
            console.error("Failed to resize window", e);
        }
    }
}
