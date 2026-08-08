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
    if (!modal) return;
    modal.classList.add('is-open');
    modal.classList.remove('hidden');
    modal.classList.add('flex');
  }

  function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;
    modal.classList.remove('is-open');
    modal.classList.add('hidden');
    modal.classList.remove('flex');
  }

  let previewObjectURL = null;

  function revokePreviewObjectURL() {
    if (previewObjectURL) {
      URL.revokeObjectURL(previewObjectURL);
      previewObjectURL = null;
    }
  }

  function setPreviewLoading(visible) {
    const loading = document.getElementById('document-preview-loading');
    if (!loading) return;
    loading.classList.toggle('is-visible', !!visible);
  }

  function hideEl(el) {
    if (!el) return;
    el.hidden = true;
    el.classList.add('hidden');
  }

  function showEl(el) {
    if (!el) return;
    el.hidden = false;
    el.classList.remove('hidden');
  }

  function resetPreviewMedia() {
    const iframe = document.getElementById('document-preview-iframe');
    const embed = document.getElementById('document-preview-embed');
    const image = document.getElementById('document-preview-image');
    const unsupported = document.getElementById('document-preview-unsupported');
    revokePreviewObjectURL();
    if (iframe) {
      hideEl(iframe);
      iframe.removeAttribute('src');
    }
    if (embed) {
      hideEl(embed);
      embed.removeAttribute('src');
    }
    if (image) {
      hideEl(image);
      image.removeAttribute('src');
    }
    hideEl(unsupported);
  }

  function isFirefox() {
    return typeof navigator !== 'undefined' && /firefox/i.test(navigator.userAgent);
  }

  /**
   * Abre vista previa de adjunto (PDF / imagen).
   * Usa fetch+blob para enviar cookie de sesión; en Firefox PDF usa <embed>.
   */
  async function openDocumentPreview(url, filename, kind) {
    const modal = document.getElementById('document-preview-modal');
    if (!modal || !url) {
      if (url) window.open(url, '_blank', 'noopener,noreferrer');
      return;
    }

    const titleEl = document.getElementById('document-preview-filename');
    const openLink = document.getElementById('document-preview-open');
    const iframe = document.getElementById('document-preview-iframe');
    const embed = document.getElementById('document-preview-embed');
    const image = document.getElementById('document-preview-image');
    const unsupported = document.getElementById('document-preview-unsupported');

    if (titleEl) titleEl.textContent = filename || '';
    if (openLink) openLink.href = url;

    resetPreviewMedia();
    setPreviewLoading(true);
    openModal('document-preview-modal');

    try {
      const response = await fetch(url, {
        method: 'GET',
        credentials: 'same-origin',
        headers: { Accept: '*/*' },
      });
      if (!response.ok) {
        throw new Error('preview_http_' + response.status);
      }

      const rawBlob = await response.blob();
      const headerType = (response.headers.get('content-type') || '').split(';')[0].trim().toLowerCase();
      const blobType = (rawBlob.type || headerType || '').toLowerCase();
      const lowerName = (filename || '').toLowerCase();

      let resolvedKind = kind;
      if (!resolvedKind || resolvedKind === 'other') {
        if (blobType.includes('pdf') || lowerName.endsWith('.pdf')) {
          resolvedKind = 'pdf';
        } else if (blobType.startsWith('image/') || /\.(png|jpe?g|gif|webp)$/.test(lowerName)) {
          resolvedKind = 'image';
        }
      }

      let mime = blobType;
      if (resolvedKind === 'pdf' && !mime.includes('pdf')) {
        mime = 'application/pdf';
      } else if (resolvedKind === 'image' && !mime.startsWith('image/')) {
        if (lowerName.endsWith('.png')) mime = 'image/png';
        else if (lowerName.endsWith('.webp')) mime = 'image/webp';
        else if (lowerName.endsWith('.gif')) mime = 'image/gif';
        else mime = 'image/jpeg';
      }

      const typedBlob = mime && mime !== rawBlob.type ? new Blob([rawBlob], { type: mime }) : rawBlob;
      previewObjectURL = URL.createObjectURL(typedBlob);

      if (resolvedKind === 'pdf') {
        // Firefox suele manejar mejor PDF con <embed> + MIME tipado.
        if (isFirefox() && embed) {
          embed.type = 'application/pdf';
          embed.src = previewObjectURL;
          showEl(embed);
        } else if (iframe) {
          iframe.src = previewObjectURL;
          showEl(iframe);
        } else if (embed) {
          embed.type = 'application/pdf';
          embed.src = previewObjectURL;
          showEl(embed);
        }
      } else if (resolvedKind === 'image' && image) {
        image.src = previewObjectURL;
        image.alt = filename || 'Vista previa';
        showEl(image);
      } else {
        showEl(unsupported);
      }
    } catch (err) {
      showEl(unsupported);
    } finally {
      setPreviewLoading(false);
    }
  }

  function closeDocumentPreview() {
    resetPreviewMedia();
    setPreviewLoading(false);
    closeModal('document-preview-modal');
  }

  function initHtmxAuth() {
    document.body.addEventListener('htmx:configRequest', (event) => {
      const token = sessionStorage.getItem('access_token');
      if (token) {
        event.detail.headers['Authorization'] = 'Bearer ' + token;
      }
    });
  }

  function setFieldIfAllowed(id, value, emptyOnly) {
    if (value == null || value === '') return;
    const el = document.getElementById(id);
    if (!el) return;
    const current = (el.value || '').trim();
    if (emptyOnly && current !== '') return;
    el.value = value;
  }

  function readExtraFields() {
    const input = document.getElementById('extra_fields');
    if (!input) return [];
    try {
      const parsed = JSON.parse(input.value || '[]');
      return Array.isArray(parsed) ? parsed : [];
    } catch (_) {
      return [];
    }
  }

  function writeExtraFields(fields) {
    const input = document.getElementById('extra_fields');
    const list = document.getElementById('extra-fields-badges');
    const empty = document.getElementById('extra-fields-empty');
    if (!input || !list) return;

    const safe = Array.isArray(fields) ? fields.filter((f) => f && f.key && f.label && f.value) : [];
    input.value = JSON.stringify(safe);
    list.innerHTML = '';
    safe.forEach((f) => {
      const badge = document.createElement('span');
      badge.className = 'extra-field-badge';
      badge.dataset.key = f.key;
      badge.dataset.label = f.label;
      badge.dataset.value = f.value;

      const text = document.createElement('span');
      text.className = 'extra-field-badge-text';
      text.textContent = f.label + ': ' + f.value;

      const btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'extra-field-badge-remove';
      btn.setAttribute('aria-label', 'Quitar ' + f.label);
      btn.dataset.removeExtra = '1';
      btn.textContent = '×';

      badge.appendChild(text);
      badge.appendChild(btn);
      list.appendChild(badge);
    });
    if (empty) empty.classList.toggle('hidden', safe.length > 0);
  }

  function mergeExtraFields(incoming, replaceAll) {
    const current = readExtraFields();
    const next = replaceAll ? [] : current.slice();
    const seen = new Set(next.map((f) => String(f.key || '').toLowerCase()));
    (incoming || []).forEach((f) => {
      if (!f || !f.key || !f.label || !f.value) return;
      const key = String(f.key).toLowerCase();
      if (seen.has(key)) return;
      seen.add(key);
      next.push({ key: f.key, label: f.label, value: String(f.value) });
    });
    writeExtraFields(next);
  }

  function initExtraFieldBadges() {
    const list = document.getElementById('extra-fields-badges');
    if (!list) return;
    // Sync from server-rendered badges / hidden input.
    writeExtraFields(readExtraFields());
    list.addEventListener('click', (e) => {
      const btn = e.target.closest('[data-remove-extra]');
      if (!btn) return;
      const badge = btn.closest('.extra-field-badge');
      if (!badge) return;
      const key = badge.dataset.key;
      writeExtraFields(readExtraFields().filter((f) => f.key !== key));
    });
  }

  function applyOCRSuggestions(data, emptyOnly) {
    if (!data || typeof data !== 'object') return;
    setFieldIfAllowed('title', data.title, emptyOnly);
    setFieldIfAllowed('issuer', data.issuer, emptyOnly);
    setFieldIfAllowed('reference_number', data.reference_number, emptyOnly);
    setFieldIfAllowed('document_date', data.document_date, emptyOnly);
    setFieldIfAllowed('due_date', data.due_date, emptyOnly);
    setFieldIfAllowed('amount', data.amount, emptyOnly);
    setFieldIfAllowed('currency', data.currency, emptyOnly);
    // Prefer structured extras over free-text notes from OCR.
    const extras = Array.isArray(data.extra_fields)
      ? data.extra_fields
      : Array.isArray(data.extraFields)
        ? data.extraFields
        : [];
    if (extras.length) {
      mergeExtraFields(extras, !emptyOnly);
      const notesEl = document.getElementById('notes');
      if (notesEl && /^Total sugerido:/i.test((notesEl.value || '').trim())) {
        notesEl.value = '';
      }
    } else if (data.notes) {
      setFieldIfAllowed('notes', data.notes, emptyOnly);
    }
  }

  async function analyzeDocumentWithOCR(btn) {
    const statusEls = document.querySelectorAll('#ocr-status');
    const setStatus = (msg) => statusEls.forEach((el) => { el.textContent = msg; });
    const inputId = btn?.dataset?.ocrFileInput || 'file';
    const emptyOnly = btn?.dataset?.ocrFillEmptyOnly === 'true';
    const input = document.getElementById(inputId);
    const file = input?.files?.[0];

    if (!file) {
      setStatus('Selecciona un archivo primero.');
      return;
    }

    setStatus('Analizando con OCR…');
    btn.disabled = true;

    try {
      const body = new FormData();
      body.append('file', file);
      const res = await fetch('/archivo/documentos/ocr-sugerir', {
        method: 'POST',
        body,
        credentials: 'same-origin',
        headers: { Accept: 'application/json' },
      });
      const payload = await res.json().catch(() => ({}));
      if (!res.ok) {
        const msg =
          payload?.error ||
          payload?.message ||
          'No se pudo analizar el archivo.';
        setStatus(msg);
        return;
      }
      const data = payload?.data ?? payload;
      applyOCRSuggestions(data, emptyOnly);
      const n = Array.isArray(data?.extra_fields)
        ? data.extra_fields.length
        : Array.isArray(data?.extraFields)
          ? data.extraFields.length
          : 0;
      setStatus(
        n > 0
          ? 'Campos y ' + n + ' dato(s) adicional(es) sugeridos. Revísalos antes de guardar.'
          : 'Campos sugeridos. Revísalos antes de guardar.',
      );
    } catch (_) {
      setStatus('Error de red al analizar el archivo.');
    } finally {
      btn.disabled = false;
    }
  }

  function initOCRSuggest() {
    const btn = document.getElementById('ocr-analyze-btn');
    if (!btn) return;
    btn.addEventListener('click', () => {
      analyzeDocumentWithOCR(btn);
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

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeDocumentPreview();
  });

  document.addEventListener('DOMContentLoaded', () => {
    initSidebar();
    initTheme();
    initHtmxAuth();
    initOCRSuggest();
    initExtraFieldBadges();
  });

  window.togglePassword = togglePassword;
  window.toggleSidebar = toggleSidebar;
  window.toggleProfileMenu = toggleProfileMenu;
  window.toggleMobileMenu = toggleMobileMenu;
  window.setThemeMode = setThemeMode;
  window.setThemeColor = setThemeColor;
  window.openModal = openModal;
  window.closeModal = closeModal;
  window.openDocumentPreview = openDocumentPreview;
  window.closeDocumentPreview = closeDocumentPreview;
})();
