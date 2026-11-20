(() => {
  'use strict';
  const $ = id => document.getElementById(id);
  const source = $('source'), file = $('file'), language = $('language');
  const limit = 256 * 1024;
  const supported = typeof fetch === 'function' && typeof Blob === 'function';
  const local = { mode: 'same-origin', credentials: 'omit', redirect: 'error', cache: 'no-store' };
  const flags = ['keepDoc', 'keepBlank', 'keepLicense', 'replace', 'commentsOnly', 'keepComments'];
  const detectionLabels = new Map([
    ['filename', '自动·文件名'], ['shebang', '自动·首行解释器'], ['content', '自动·代码特征'], ['manual', '手动指定']
  ]);
  let filename = '', raw = null, result = null;
  let fileVersion = 0, version = 0, reading = false, busy = false, controller = null;

  function say(message, error = false) {
    $('status').textContent = message;
    $('status').classList.toggle('error', error);
  }

  function refresh() {
    // A textarea normalizes CR/CRLF; only a real edit should discard the file's original bytes.
    const text = raw === null ? source.value : raw;
    const size = text.length > limit ? limit + 1 : (supported ? new Blob([text]).size : 0);
    const error = size > limit ? '输入超过 256 KiB，请缩短代码或选择较小的文件。' :
      text.includes('\0') ? '输入含 NUL 字符，请使用 UTF-8 文本。' :
      /[\uD800-\uDFFF]/u.test(text) ? '输入含不完整的 Unicode 字符。' : '';
    $('input-size').textContent = size > limit ? '超过 256 KiB' : `${size.toLocaleString()} B`;
    $('input-error').textContent = error;
    source.setAttribute('aria-invalid', String(Boolean(error)));
    $('clear').disabled = !reading && !filename && !text && language.value === 'auto';
    $('preview').disabled = !supported || busy || reading || Boolean(error) || (!filename && !text);
  }

  function invalidate(message = '') {
    version++;
    if (controller) controller.abort();
    controller = null;
    busy = false;
    result = null;
    $('empty').hidden = false;
    $('output').hidden = $('counts').hidden = $('changes').hidden = $('diff-note').hidden = $('jump').hidden = true;
    $('output').textContent = $('counts').textContent = $('diff').textContent = $('diff-note').textContent = '';
    $('changes').open = false;
    $('copy').disabled = $('download').disabled = true;
    $('result-state').textContent = '待预览';
    say(message);
    refresh();
  }

  function setFilename(name) {
    filename = name;
    $('filename').textContent = name || '粘贴输入';
    $('filename').title = name;
  }

  file.addEventListener('change', async () => {
    const selected = file.files[0];
    if (!selected) return;
    file.value = ''; // Allow picking the same file again after an edit or error.
    const ticket = ++fileVersion;
    reading = true;
    invalidate('正在本机读取文件，尚未发送。');
    try {
      if (selected.size > limit) throw new Error('文件超过 256 KiB。');
      const bytes = selected.arrayBuffer ? await selected.arrayBuffer() : await new Response(selected).arrayBuffer();
      if (ticket !== fileVersion) return;
      let text;
      try { text = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(bytes); }
      catch { throw new Error('文件不是有效的 UTF-8 文本。'); }
      if (text.includes('\0')) throw new Error('文件含 NUL 字符。');
      const name = selected.name.split(/[\\/]/).pop();
      if (!name) throw new Error('无法读取文件名。');
      setFilename(name);
      raw = text;
      source.value = text;
      say('文件已载入，尚未发送。');
    } catch (error) {
      if (ticket === fileVersion) say(`未载入文件，原输入未变。${error.message}`, true);
    } finally {
      if (ticket === fileVersion) { reading = false; refresh(); }
    }
  });

  source.addEventListener('input', () => {
    fileVersion++;
    reading = false;
    raw = null;
    invalidate();
  });
  $('clear').addEventListener('click', () => {
    fileVersion++;
    reading = false;
    raw = null;
    file.value = source.value = '';
    setFilename('');
    language.value = 'auto';
    invalidate('输入和结果已清空。');
    source.focus();
  });
  function syncGlossary() {
    const enabled = $('replace').checked;
    $('term-options').disabled = !enabled;
    if (!enabled) $('commentsOnly').checked = $('keepComments').checked = false;
  }
  language.addEventListener('change', () => invalidate());
  for (const id of flags) $(id).addEventListener('change', () => { syncGlossary(); invalidate(); });

  async function loadLanguages() {
    try {
      const response = await fetch('api/languages', local);
      if (!response.ok) throw new Error();
      const entries = await response.json();
      if (!Array.isArray(entries) || !entries.length) throw new Error();
      const options = document.createDocumentFragment();
      for (const entry of entries) {
        if (!entry || typeof entry.name !== 'string' || typeof entry.risky !== 'boolean') throw new Error();
        if (entry.name !== 'go' && entry.name !== 'auto') {
          options.append(new Option(entry.name + (entry.risky ? '（启发式）' : ''), entry.name));
        }
      }
      language.append(options); // Never reset a language chosen while the list was loading.
    } catch {
      $('language-help').textContent = '未能载入完整语言列表。仍可自动识别或手动选择 Go；刷新页面可重试。';
    }
  }

  function render(data) {
    result = data;
    $('empty').hidden = true;
    $('output').textContent = data.output;
    $('output').hidden = $('counts').hidden = $('jump').hidden = false;
    $('result-state').textContent = (data.changed ? '有改动' : '无改动') + (data.output === '' ? ' · 空输出' : '');
    $('counts').textContent = `${data.language}（${detectionLabels.get(data.detectedBy)}） · 移除 ${data.comments} 条注释、${data.removedLines} 行；替换 ${data.terms} 处术语。`;
    $('diff').textContent = data.diffOmitted ? '' : data.diff;
    $('changes').hidden = data.diffOmitted || !data.diff;
    $('diff-note').hidden = !$('changes').hidden;
    $('diff-note').textContent = data.diffOmitted ? '差异较大，已省略逐行显示。仍可复制或下载完整结果。' : '没有可显示的逐行差异。';
    $('copy').disabled = false;
    $('download').disabled = typeof URL.createObjectURL !== 'function';
    say(`预览已生成：${data.language}（${detectionLabels.get(data.detectedBy)}）。原文件未变。`);
  }

  $('workspace').addEventListener('submit', async event => {
    event.preventDefault();
    if ($('preview').disabled) return;
    invalidate('正在生成预览…');
    const ticket = version;
    busy = true;
    $('preview').disabled = true;
    controller = typeof AbortController === 'function' ? new AbortController() : null;
    const body = { source: raw === null ? source.value : raw, language: language.value };
    if (filename) body.filename = filename;
    for (const id of flags) body[id] = $(id).checked;
    try {
      const response = await fetch('api/preview', {
        ...local, method: 'POST', headers: { 'Content-Type': 'application/json; charset=utf-8' },
        body: JSON.stringify(body), signal: controller ? controller.signal : undefined
      });
      const data = await response.json().catch(() => null);
      if (ticket !== version) return;
      if (!response.ok) {
        const detail = data && typeof data.error === 'string' ? data.error : `HTTP ${response.status}`;
        say(`未生成预览：${detail}`, true);
      } else if (!data || typeof data.output !== 'string' || typeof data.diff !== 'string' ||
          typeof data.language !== 'string' || !detectionLabels.has(data.detectedBy) ||
          typeof data.changed !== 'boolean' || typeof data.diffOmitted !== 'boolean' ||
          ![data.comments, data.removedLines, data.terms].every(n => Number.isSafeInteger(n) && n >= 0)) {
        say('本机服务返回了无法读取的结果。', true);
      } else render(data);
    } catch {
      if (ticket === version) say('无法连接本机预览服务，请重试。', true);
    } finally {
      if (ticket === version) { busy = false; controller = null; refresh(); }
    }
  });

  // Older browsers can copy via the native copy event, including a valid empty result.
  function copyFallback(text) {
    let copied = false;
    const transfer = event => {
      if (!event.clipboardData) return;
      event.clipboardData.setData('text/plain', text);
      event.preventDefault();
      copied = true;
    };
    document.addEventListener('copy', transfer);
    try { return document.execCommand('copy') && copied; }
    finally { document.removeEventListener('copy', transfer); }
  }
  $('copy').addEventListener('click', async () => {
    const snapshot = result;
    if (!snapshot) return;
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) await navigator.clipboard.writeText(snapshot.output);
      else if (!copyFallback(snapshot.output)) throw new Error();
      if (result === snapshot) say(snapshot.output === '' ? '已复制空结果。' : '已复制结果。');
    } catch { if (result === snapshot) say('无法写入剪贴板，可手动复制或下载副本。', true); }
  });
  $('download').addEventListener('click', () => {
    if (!result) return;
    const link = document.createElement('a');
    const url = URL.createObjectURL(new Blob([result.output], { type: 'application/octet-stream' }));
    link.href = url;
    link.download = 'preview-' + (filename || 'source.txt');
    link.hidden = true;
    document.body.append(link);
    try { link.click(); say('已请求下载预览副本。'); }
    finally { link.remove(); setTimeout(() => URL.revokeObjectURL(url), 1000); }
  });

  syncGlossary();
  refresh();
  if (supported) loadLanguages();
  else say('当前浏览器不支持本机预览，请使用较新的浏览器。', true);
  if (typeof TextDecoder !== 'function') {
    file.disabled = true;
    $('source-help').textContent = '此浏览器不能验证文件的 UTF-8 编码，请改用粘贴输入。';
  }
})();
