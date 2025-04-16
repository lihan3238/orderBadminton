function updateStatus() {
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            const statusDiv = document.getElementById('status');
            const listDiv = document.getElementById('roomList');

            const today = data.today_available || [];
            const tomorrow = data.tomorrow_available || [];

            if (today.length === 0 && tomorrow.length === 0) {
                statusDiv.textContent = '❌ 场地已被预约';
                statusDiv.style.color = 'red';
                listDiv.innerHTML = '';
                return;
            }

            statusDiv.textContent = '✅ 有空闲场地';
            statusDiv.style.color = 'green';

            let html = '';
            if (today.length > 0) {
                html += '<strong>今天空闲场地：</strong><ul>' +
                    today.map(room => `<li>${room}</li>`).join('') +
                    '</ul>';
            }
            if (tomorrow.length > 0) {
                html += '<strong>明天空闲场地：</strong><ul>' +
                    tomorrow.map(room => `<li>${room}</li>`).join('') +
                    '</ul>';
            }
            listDiv.innerHTML = html;
        })
        .catch(error => {
            console.error('获取状态失败:', error);
        });
}

// 初始加载
updateStatus();
// 每隔 15 秒更新一次状态
setInterval(updateStatus, 15000);
