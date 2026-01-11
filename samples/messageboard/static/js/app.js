/**
 * 留言板前端交互逻辑
 */

(function () {
  "use strict";

  // API 基础路径
  const API_BASE = "/api";

  // 初始化
  $(document).ready(function () {
    loadMessages();
    setupFormValidation();
  });

  /**
   * 加载留言列表
   */
  function loadMessages() {
    $.ajax({
      url: `${API_BASE}/messages`,
      method: "GET",
      success: function (response) {
        if (response.code === 200) {
          renderMessages(response.data);
        } else {
          showError("加载留言失败");
        }
      },
      error: function () {
        showError("网络错误，请稍后重试");
      },
    });
  }

  /**
   * 渲染留言列表
   */
  function renderMessages(messages) {
    const $list = $("#message-list");

    if (!messages || messages.length === 0) {
      $list.html(`
                <div class="empty-state">
                    <div class="empty-state-icon">💬</div>
                    <p>还没有留言，快来发表第一条吧！</p>
                </div>
            `);
      return;
    }

    const html = messages
      .map(function (msg) {
        const date = new Date(msg.created_at);
        const timeStr = formatDate(date);

        return `
                <div class="message-item">
                    <div class="message-header">
                        <span class="message-nickname">${escapeHtml(msg.nickname)}</span>
                        <span class="message-time">${timeStr}</span>
                    </div>
                    <div class="message-content">${escapeHtml(msg.content)}</div>
                </div>
            `;
      })
      .join("");

    $list.html(html);
  }

  /**
   * 设置表单验证
   */
  function setupFormValidation() {
    $("#message-form").validate({
      rules: {
        nickname: {
          required: true,
          minlength: 2,
          maxlength: 20,
        },
        content: {
          required: true,
          minlength: 5,
          maxlength: 500,
        },
      },
      messages: {
        nickname: {
          required: "请输入昵称",
          minlength: "昵称至少需要2个字符",
          maxlength: "昵称不能超过20个字符",
        },
        content: {
          required: "请输入留言内容",
          minlength: "留言内容至少需要5个字符",
          maxlength: "留言内容不能超过500个字符",
        },
      },
      submitHandler: function (form) {
        submitMessage();
        return false;
      },
    });
  }

  /**
   * 提交留言
   */
  function submitMessage() {
    const data = {
      nickname: $("#nickname").val().trim(),
      content: $("#content").val().trim(),
    };

    $.ajax({
      url: `${API_BASE}/messages`,
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify(data),
      success: function (response) {
        if (response.code === 200) {
          showSuccess("留言提交成功，等待审核");
          $("#message-form")[0].reset();
          // 不立即刷新列表，因为留言需要审核
        } else {
          showError(response.message || "提交失败");
        }
      },
      error: function (xhr) {
        const response = xhr.responseJSON;
        showError(response?.message || "网络错误，请稍后重试");
      },
    });
  }

  /**
   * 格式化日期
   */
  function formatDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    return `${year}-${month}-${day} ${hours}:${minutes}`;
  }

  /**
   * HTML 转义
   */
  function escapeHtml(text) {
    const map = {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;",
    };
    return text.replace(/[&<>"']/g, function (m) {
      return map[m];
    });
  }

  /**
   * 显示成功提示
   */
  function showSuccess(message) {
    showToast(message, "success");
  }

  /**
   * 显示错误提示
   */
  function showError(message) {
    showToast(message, "error");
  }

  /**
   * 显示 Toast 提示
   */
  function showToast(message, type) {
    const className = type === "success" ? "toast-success" : "toast-error";

    const toast = $(`
            <div class="toast ${className} show">
                <span class="toast-message">${message}</span>
                <button class="toast-close">&times;</button>
            </div>
        `);

    let container = $(".toast-container");
    if (container.length === 0) {
      container = $('<div class="toast-container"></div>').appendTo("body");
    }

    container.append(toast);

    let timer = setTimeout(function () {
      toast.fadeOut(function () {
        toast.remove();
      });
    }, 3000);

    // 点击关闭按钮
    toast.find(".toast-close").on("click", function () {
      clearTimeout(timer);
      toast.fadeOut(function () {
        toast.remove();
      });
    });

    // 点击 toast 本体也可以关闭
    toast.on("click", function (e) {
      if (!$(e.target).hasClass(".toast-close")) {
        clearTimeout(timer);
        toast.fadeOut(function () {
          toast.remove();
        });
      }
    });

    // 鼠标悬停时暂停自动消失
    toast.on("mouseenter", function () {
      clearTimeout(timer);
    });

    // 鼠标离开后继续计时（1秒后消失）
    toast.on("mouseleave", function () {
      timer = setTimeout(function () {
        toast.fadeOut(function () {
          toast.remove();
        });
      }, 1000);
    });
  }
})();
