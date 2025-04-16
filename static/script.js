function updateStatus() {
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            const statusDiv = document.getElementById('status');
            const listDiv = document.getElementById('roomList');

            if (data.available) {
                statusDiv.textContent = '✅ 有空闲场地';
                statusDiv.style.color = 'green';

                // 显示空闲场地名称列表
                listDiv.innerHTML = '<strong>空闲场地：</strong><ul>' +
                    data.rooms.map(room => `<li>${room}</li>`).join('') +
                    '</ul>';
            } else {
                statusDiv.textContent = '❌ 场地已被预约';
                statusDiv.style.color = 'red';
                listDiv.innerHTML = '';
            }
        })
        .catch(error => {
            console.error('获取状态失败:', error);
        });
}

// 初始加载
updateStatus();

// 每隔 30 秒更新一次状态
setInterval(updateStatus, 30000);
