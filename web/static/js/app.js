/**
 * Utilidades vanilla JS para docsy HTMX frontend.
 * Sin dependencias externas (excepto htmx.min.js).
 */

(function () {
  'use strict';

  const SIDEBAR_KEY = 'docsy-sidebar-expanded';
  const THEME_MODE_KEY = 'docsy-theme-mode';
  const THEME_COLOR_KEY = 'docsy-theme-color';
  const ACTIVE_CLASSES = ['border-muted-foreground/30', 'bg-card'];

  /**
   * Alterna visibilidad de contraseña.
   * @param {string} inputId - ID del input password
   * @param {HTMLElement} buttonEl - Elemento clickeado (contiene img.password-toggle-icon)
   */
  function togglePassword(inputId, buttonEl) {
    const input = document.getElementById(inputId);
    if (!input) return;

    const icon = buttonEl.querySelector('.password-toggle-icon') || buttonEl.querySelector('img');
    const isPassword = input.type === 'password';
    input.type = isPassword ? 'text' : 'password';

    if (icon && icon.dataset.visibleIcon && icon.dataset.hiddenIcon) {
      icon.src = isPassword ? icon.dataset.visibleIcon : icon.dataset.hiddenIcon;
    }
  }

  /**
   * Colapsa o expande el sidebar (desktop).
   * Persiste estado en localStorage.
   */
  function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    if (!sidebar) return;

    sidebar.classList.toggle('sidebar-collapsed');
    const isCollapsed = sidebar.classList.contains('sidebar-collapsed');

    sidebar.classList.toggle('w-[210px]', !isCollapsed);
    sidebar.classList.toggle('xl:w-[280px]', !isCollapsed);
    sidebar.classList.toggle('w-[70px]', isCollapsed);

    sidebar.querySelector('.sidebar-toggle')?.classList.toggle('rotate-180', isCollapsed);
    sidebar.querySelectorAll('.sidebar-item-label, .sidebar-group-label, .sidebar-version').forEach((el) => {
      el.classList.toggle('hidden', isCollapsed);
    });
    sidebar.querySelector('.sidebar-logo')?.classList.toggle('hidden', isCollapsed);

    localStorage.setItem(SIDEBAR_KEY, isCollapsed ? '0' : '1');
  }

  function initSidebar() {
    const sidebar = document.getElementById('sidebar');
    if (!sidebar) return;

    const stored = localStorage.getItem(SIDEBAR_KEY);
    if (stored === '0') {
      sidebar.classList.add('sidebar-collapsed', 'w-[70px]');
      sidebar.classList.remove('w-[210px]', 'xl:w-[280px]');
      sidebar.querySelector('.sidebar-toggle')?.classList.add('rotate-180');
      sidebar.querySelectorAll('.sidebar-item-label, .sidebar-group-label, .sidebar-version').forEach((el) => {
        el.classList.add('hidden');
      });
      sidebar.querySelector('.sidebar-logo')?.classList.add('hidden');
    }
  }

  function toggleProfileMenu() {
    const menu = document.getElementById('profile-menu');
    const button = document.getElementById('profile-menu-button');
    if (!menu) return;

    const isHidden = menu.classList.toggle('hidden');
    menu.classList.toggle('scale-100', !isHidden);
    menu.classList.toggle('opacity-100', !isHidden);
    if (button) {
      button.setAttribute('aria-expanded', isHidden ? 'false' : 'true');
    }
  }

  function toggleMobileMenu() {
    document.getElementById('mobile-menu')?.classList.toggle('hidden');
  }

  function clearActiveState(selector) {
    document.querySelectorAll(selector).forEach((btn) => {
      btn.classList.remove(...ACTIVE_CLASSES);
    });
  }

  function markActiveButton(selector) {
    const active = document.querySelector(selector);
    active?.classList.add(...ACTIVE_CLASSES);
  }

  function setThemeMode(mode) {
    const root = document.documentElement;
    if (mode === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
    localStorage.setItem(THEME_MODE_KEY, mode);
    updateThemeModeButtons(mode);
  }

  function setThemeColor(color) {
    const root = document.documentElement;
    if (color) {
      root.setAttribute('data-theme', color);
    } else {
      root.removeAttribute('data-theme');
    }
    localStorage.setItem(THEME_COLOR_KEY, color);
    updateThemeColorButtons(color);
  }

  function updateThemeModeButtons(mode) {
    clearActiveState('.theme-mode-light, .theme-mode-dark');
    markActiveButton(mode === 'dark' ? '.theme-mode-dark' : '.theme-mode-light');
  }

  function updateThemeColorButtons(color) {
    clearActiveState('[data-theme-color]');
    const selector = color === ''
      ? '[data-theme-color=""]'
      : `[data-theme-color="${color}"]`;
    markActiveButton(selector);
  }

  function initTheme() {
    const mode = localStorage.getItem(THEME_MODE_KEY) || 'light';
    const color = localStorage.getItem(THEME_COLOR_KEY) || '';

    setThemeMode(mode);
    setThemeColor(color);
  }

  function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) modal.classList.remove('hidden');
  }

  function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) modal.classList.add('hidden');
  }

  function initHtmxAuth() {
    document.body.addEventListener('htmx:configRequest', (event) => {
      const token = sessionStorage.getItem('access_token');
      if (token) {
        event.detail.headers['Authorization'] = 'Bearer ' + token;
      }
    });
  }

  document.addEventListener('click', (e) => {
    const profileBtn = document.getElementById('profile-menu-button');
    const profileMenu = document.getElementById('profile-menu');
    if (profileMenu && profileBtn && !profileBtn.contains(e.target) && !profileMenu.contains(e.target)) {
      profileMenu.classList.add('hidden');
      profileBtn.setAttribute('aria-expanded', 'false');
    }
  });

  document.addEventListener('DOMContentLoaded', () => {
    initSidebar();
    initTheme();
    initHtmxAuth();
  });

  window.togglePassword = togglePassword;
  window.toggleSidebar = toggleSidebar;
  window.toggleProfileMenu = toggleProfileMenu;
  window.toggleMobileMenu = toggleMobileMenu;
  window.setThemeMode = setThemeMode;
  window.setThemeColor = setThemeColor;
  window.openModal = openModal;
  window.closeModal = closeModal;
})();
