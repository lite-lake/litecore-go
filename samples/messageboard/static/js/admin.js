/**
 * 留言板管理后台前端交互逻辑
 */

(function() {
    'use strict';

    // API 基础路径
    const API_BASE = '/api';

    // 存储 Token
    let authToken = localStorage.getItem('admin_token') || null;

    // 初始化
    $(document).ready(function() {
        checkLoginStatus();
        setupEventHandlers();
    });

    /**
     * 检查登录状态
     */
    function checkLoginStatus() {
        if (authToken) {
            showAdminPanel();
        } else {
            showLoginPanel();
        }
    }

    /**
     * 设置事件处理
     */
    function setupEventHandlers() {
        // 登录表单
        $('#login-form').on('submit', function(e) {
            e.preventDefault();
            login();
        });

        // 退出登录
        $('#logout-btn').on('click', function() {
            logout();
        });
    }

    /**
     * 登录
     */
    function login() {
        const password = $('#password').val();

        $.ajax({
            url: `${API_BASE}/admin/login`,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ password: password }),
            success: function(response) {
                if (response.code === 200 && response.data && response.data.token) {
                    authToken = response.data.token;
                    localStorage.setItem('admin_token', authToken);
                    showSuccess('登录成功');
                    showAdminPanel();
                } else {
                    showError('登录失败');
                }
            },
            error: function(xhr) {
                const response = xhr.responseJSON;
                showError(response?.message || '登录失败');
            },
            complete: function() {
                $('#password').val('');
            }
        });
    }

    /**
     * 退出登录
     */
    function logout() {
        localStorage.removeItem('admin_token');
        authToken = null;
        showLoginPanel();
        showSuccess('已退出登录');
    }

    /**
     * 显示登录面板
     */
    function showLoginPanel() {
        $('#login-panel').removeClass('d-none');
        $('#admin-panel').addClass('d-none');
    }

    /**
     * 显示管理面板
     */
    function showAdminPanel() {
        $('#login-panel').addClass('d-none');
        $('#admin-panel').removeClass('d-none');
        loadMessages();
        loadStatistics();
    }

    /**
     * 加载留言列表
     */
    function loadMessages() {
        $.ajax({
            url: `${API_BASE}/admin/messages`,
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            },
            success: function(response) {
                if (response.code === 200 && response.data) {
                    renderMessages(response.data);
                } else {
                    showError('加载留言失败');
                }
            },
            error: function(xhr) {
                if (xhr.status === 401) {
                    logout();
                    showError('登录已过期，请重新登录');
                } else {
                    showError('加载留言失败');
                }
            }
        });
    }

    /**
     * 渲染留言列表
     */
    function renderMessages(messages) {
        const $list = $('#admin-message-list');

        if (!messages || messages.length === 0) {
            $list.html(`
                <tr>
                    <td colspan="5" class="text-center py-5 text-muted">
                        暂无留言
                    </td>
                </tr>
            `);
            return;
        }

        const html = messages.map(function(msg) {
            const date = new Date(msg.created_at);
            const timeStr = formatDate(date);
            const statusBadge = getStatusBadge(msg.status);
            const actions = getActionButtons(msg);

            return `
                <tr data-id="${msg.id}">
                    <td>${escapeHtml(msg.nickname)}</td>
                    <td>${escapeHtml(msg.content)}</td>
                    <td>${statusBadge}</td>
                    <td class="text-muted">${timeStr}</td>
                    <td>${actions}</td>
                </tr>
            `;
        }).join('');

        $list.html(html);
    }

    /**
     * 获取状态标签
     */
    function getStatusBadge(status) {
        const badges = {
            'pending': '<span class="badge badge-pending">待审核</span>',
            'approved': '<span class="badge badge-approved">已通过</span>',
            'rejected': '<span class="badge badge-rejected">已拒绝</span>'
        };
        return badges[status] || '<span class="badge badge-pending">未知</span>';
    }

    /**
     * 获取操作按钮
     */
    function getActionButtons(msg) {
        let buttons = '';

        if (msg.status === 'pending') {
            buttons += `
                <button class="btn btn-sm btn-success me-1" onclick="approveMessage(${msg.id})">通过</button>
                <button class="btn btn-sm btn-danger me-1" onclick="rejectMessage(${msg.id})">拒绝</button>
            `;
        }

        buttons += `
            <button class="btn btn-sm btn-outline-secondary" onclick="deleteMessage(${msg.id})">删除</button>
        `;

        return buttons;
    }

    /**
     * 通过留言
     */
    window.approveMessage = function(id) {
        updateStatus(id, 'approved');
    };

    /**
     * 拒绝留言
     */
    window.rejectMessage = function(id) {
        updateStatus(id, 'rejected');
    };

    /**
     * 更新留言状态
     */
    function updateStatus(id, status) {
        $.ajax({
            url: `${API_BASE}/admin/messages/${id}/status`,
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`
            },
            data: { status: status },
            success: function(response) {
                if (response.code === 200) {
                    showSuccess('状态更新成功');
                    loadMessages();
                    loadStatistics();
                } else {
                    showError(response.message || '更新失败');
                }
            },
            error: function(xhr) {
                const response = xhr.responseJSON;
                showError(response?.message || '更新失败');
            }
        });
    }

    /**
     * 删除留言
     */
    window.deleteMessage = function(id) {
        if (!confirm('确定要删除这条留言吗？')) {
            return;
        }

        $.ajax({
            url: `${API_BASE}/admin/messages/${id}/delete`,
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authToken}`
            },
            success: function(response) {
                if (response.code === 200) {
                    showSuccess('删除成功');
                    loadMessages();
                    loadStatistics();
                } else {
                    showError(response.message || '删除失败');
                }
            },
            error: function(xhr) {
                const response = xhr.responseJSON;
                showError(response?.message || '删除失败');
            }
        });
    }

    /**
     * 加载统计数据
     */
    function loadStatistics() {
        // 从留言列表中计算统计数据
        $.ajax({
            url: `${API_BASE}/admin/messages`,
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            },
            success: function(response) {
                if (response.code === 200 && response.data) {
                    const messages = response.data;
                    const stats = {
                        pending: 0,
                        approved: 0,
                        rejected: 0,
                        total: messages.length
                    };

                    messages.forEach(function(msg) {
                        if (stats[msg.status] !== undefined) {
                            stats[msg.status]++;
                        }
                    });

                    $('#pending-count').text(stats.pending);
                    $('#approved-count').text(stats.approved);
                    $('#rejected-count').text(stats.rejected);
                    $('#total-count').text(stats.total);
                }
            }
        });
    }

    /**
     * 格式化日期
     */
    function formatDate(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        return `${year}-${month}-${day} ${hours}:${minutes}`;
    }

    /**
     * HTML 转义
     */
    function escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, function(m) {
            return map[m];
        });
    }

    /**
     * 显示成功提示
     */
    function showSuccess(message) {
        showToast(message, 'success');
    }

    /**
     * 显示错误提示
     */
    function showError(message) {
        showToast(message, 'error');
    }

    /**
     * 显示 Toast 提示
     */
    function showToast(message, type) {
        const className = type === 'success' ? 'toast-success' : 'toast-error';

        const toast = $(`
            <div class="toast ${className}">
                ${message}
            </div>
        `);

        let container = $('.toast-container');
        if (container.length === 0) {
            container = $('<div class="toast-container"></div>').appendTo('body');
        }

        container.append(toast);

        setTimeout(function() {
            toast.fadeOut(function() {
                toast.remove();
            });
        }, 3000);
    }

})();
