const TEXT = navigator.language.toLowerCase().startsWith('ko') ? {
  subtitle: '빠르고 안전한 로컬 변환. 데이터는 이 PC 밖으로 나가지 않습니다.',
  localBadge: '로컬 전용', dropTitle: 'CSV 파일을 여기에 놓으세요', dropHint: '또는 클릭해서 파일 선택',
  selectedLabel: '선택한 파일', changeFile: '변경', sizeLabel: '크기', delimiterLabel: '구분자', columnsLabel: '컬럼',
  previewLabel: '미리보기', previewTitle: '첫 데이터', optionsLabel: '옵션', optionsTitle: '변환 설정',
  inferTitle: '데이터 타입 자동 추론', inferHint: '숫자와 Boolean을 JSON 타입으로 변환합니다.',
  nullTitle: '빈 셀을 null로', nullHint: '끄면 빈 문자열("")로 유지합니다.', safeTitle: '안전 변환',
  safeHint: '선행 0 값은 문자열로 보호하고 전체 CSV 성공 후에만 결과 파일을 확정합니다.',
  convertButton: 'JSON으로 변환', progressTitle: '변환 중…', progressText: 'CSV를 디스크에서 직접 스트리밍 처리하고 있습니다.',
  completeLabel: '변환 완료', completeTitle: 'JSON 파일을 만들었습니다', rowsLabel: '행', outputSizeLabel: '결과 크기',
  timeLabel: '소요 시간', revealButton: '폴더에서 보기', againButton: '다른 파일 변환',
  delimiters: { comma: '쉼표 (,)', tab: '탭', pipe: '파이프 (|)', semicolon: '세미콜론 (;)' },
  errors: {
    BACKEND_NOT_READY: '앱 백엔드가 준비되지 않았습니다. 앱을 다시 실행해 주세요.',
    FILE_READ_FAILED: '파일을 읽을 수 없습니다.', HEADER_READ_FAILED: '헤더 행을 읽을 수 없습니다.',
    EMPTY_HEADER: '빈 컬럼명이 있습니다.', DUPLICATE_HEADER: '중복 컬럼명이 있습니다.',
    MIXED_DELIMITER_DETECTED: '구분자가 혼합되었거나 판별하기 어렵습니다.', CSV_PARSE_FAILED: 'CSV 형식이 올바르지 않습니다.',
    OUTPUT_CREATE_FAILED: '결과 파일을 만들 수 없습니다.', OUTPUT_WRITE_FAILED: 'JSON 파일을 저장할 수 없습니다.',
    OUTPUT_EQUALS_INPUT: '입력 파일과 결과 파일 경로가 같을 수 없습니다.', CONVERSION_IN_PROGRESS: '이미 변환 작업이 실행 중입니다.',
    UNKNOWN_ERROR: '예상하지 못한 오류가 발생했습니다.'
  }
} : {
  subtitle: 'Fast, local conversion. Your data never leaves this computer.', localBadge: 'LOCAL ONLY',
  dropTitle: 'Drop a CSV file here', dropHint: 'or click to choose a file', selectedLabel: 'SELECTED FILE', changeFile: 'Change',
  sizeLabel: 'Size', delimiterLabel: 'Delimiter', columnsLabel: 'Columns', previewLabel: 'PREVIEW', previewTitle: 'First rows',
  optionsLabel: 'OPTIONS', optionsTitle: 'Conversion', inferTitle: 'Infer data types',
  inferHint: 'Numbers and booleans become native JSON values.', nullTitle: 'Empty cells as null',
  nullHint: 'Otherwise empty cells stay as empty strings.', safeTitle: 'Safe conversion',
  safeHint: 'Leading-zero values stay strings. Output is committed only after the full CSV succeeds.',
  convertButton: 'Convert to JSON', progressTitle: 'Converting…', progressText: 'Reading the CSV directly from disk.',
  completeLabel: 'COMPLETE', completeTitle: 'JSON file created', rowsLabel: 'Rows', outputSizeLabel: 'Output', timeLabel: 'Time',
  revealButton: 'Show in folder', againButton: 'Convert another',
  delimiters: { comma: 'Comma (,)', tab: 'Tab', pipe: 'Pipe (|)', semicolon: 'Semicolon (;)' },
  errors: {
    BACKEND_NOT_READY: 'The app backend is not ready. Restart the app.', FILE_READ_FAILED: 'The file could not be read.',
    HEADER_READ_FAILED: 'The header row could not be read.', EMPTY_HEADER: 'An empty column header was found.',
    DUPLICATE_HEADER: 'Duplicate column headers were found.', MIXED_DELIMITER_DETECTED: 'The delimiter is ambiguous or mixed.',
    CSV_PARSE_FAILED: 'The CSV is malformed.', OUTPUT_CREATE_FAILED: 'The output file could not be created.',
    OUTPUT_WRITE_FAILED: 'The JSON file could not be written.', OUTPUT_EQUALS_INPUT: 'The output path cannot be the input file.',
    CONVERSION_IN_PROGRESS: 'A conversion is already running.', UNKNOWN_ERROR: 'An unexpected error occurred.'
  }
};

let currentFile = null;
let lastOutputPath = '';
let toastTimer = null;

const $ = (id) => document.getElementById(id);

function applyText() {
  const keys = ['subtitle','localBadge','dropTitle','dropHint','selectedLabel','changeFile','sizeLabel','delimiterLabel','columnsLabel',
    'previewLabel','previewTitle','optionsLabel','optionsTitle','inferTitle','inferHint','nullTitle','nullHint','safeTitle','safeHint',
    'convertButton','progressTitle','progressText','completeLabel','completeTitle','rowsLabel','outputSizeLabel','timeLabel','revealButton','againButton'];
  for (const key of keys) if ($(key) && TEXT[key]) $(key).textContent = TEXT[key];
}

function getApp() {
  const app = window.go && window.go.main && window.go.main.App;
  if (!app) throw new Error('BACKEND_NOT_READY');
  return app;
}

function formatBytes(bytes) {
  const n = Number(bytes) || 0;
  if (n <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(Math.floor(Math.log(n) / Math.log(1024)), units.length - 1);
  const value = n / Math.pow(1024, i);
  return `${value >= 10 || i === 0 ? value.toFixed(i === 0 ? 0 : 1) : value.toFixed(2)} ${units[i]}`;
}

function formatDuration(ms) {
  const n = Number(ms) || 0;
  return n < 1000 ? `${n} ms` : `${(n / 1000).toFixed(n < 10000 ? 2 : 1)} s`;
}

function errorCode(error) {
  const raw = String((error && error.message) || error || 'UNKNOWN_ERROR');
  for (const code of Object.keys(TEXT.errors)) if (raw.includes(code)) return code;
  return 'UNKNOWN_ERROR';
}

function showToast(error) {
  const code = errorCode(error);
  const toast = $('toast');
  toast.textContent = TEXT.errors[code] || TEXT.errors.UNKNOWN_ERROR;
  toast.className = 'show';
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { toast.className = ''; }, 4000);
  console.error(error);
}

function showOnly(view) {
  for (const id of ['dropView', 'workspace', 'progressView', 'successView']) {
    $(id).classList.toggle('hidden', id !== view);
  }
}

function renderPreview(info) {
  const table = $('previewTable');
  table.replaceChildren();
  const thead = document.createElement('thead');
  const headerRow = document.createElement('tr');
  for (const header of info.headers || []) {
    const th = document.createElement('th');
    th.textContent = header;
    headerRow.appendChild(th);
  }
  thead.appendChild(headerRow);
  table.appendChild(thead);

  const tbody = document.createElement('tbody');
  for (const row of info.preview || []) {
    const tr = document.createElement('tr');
    for (let i = 0; i < (info.headers || []).length; i++) {
      const td = document.createElement('td');
      td.textContent = row[i] ?? '';
      tr.appendChild(td);
    }
    tbody.appendChild(tr);
  }
  table.appendChild(tbody);
}

function setFile(info) {
  if (!info) return;
  currentFile = info;
  $('fileName').textContent = info.name || '';
  $('filePath').textContent = info.path || '';
  $('fileSize').textContent = formatBytes(info.size);
  $('delimiter').textContent = TEXT.delimiters[info.delimiter] || info.delimiterRune || '—';
  $('columns').textContent = String(info.columns ?? '—');
  renderPreview(info);
  showOnly('workspace');
}

async function chooseFile() {
  try {
    const info = await getApp().SelectCSV();
    if (info) setFile(info);
  } catch (error) {
    showToast(error);
  }
}

async function inspectDropped(path) {
  if (!path) return;
  try {
    setFile(await getApp().InspectFile(path));
  } catch (error) {
    showToast(error);
  }
}

async function convertFile() {
  if (!currentFile) return;
  showOnly('progressView');
  $('progressBar').style.width = '0%';
  $('progressPercent').textContent = '0%';
  $('progressBytes').textContent = '0 B';
  try {
    const result = await getApp().ConvertFile(currentFile.path, {
      inferTypes: $('inferTypes').checked,
      emptyAsNull: $('emptyAsNull').checked
    });
    if (!result || result.cancelled) {
      showOnly('workspace');
      return;
    }
    lastOutputPath = result.outputPath || '';
    $('resultPath').textContent = lastOutputPath;
    $('resultRows').textContent = Number(result.rows || 0).toLocaleString();
    $('resultSize').textContent = formatBytes(result.bytes);
    $('resultTime').textContent = formatDuration(result.durationMs);
    showOnly('successView');
  } catch (error) {
    showOnly('workspace');
    showToast(error);
  }
}

function registerRuntimeHooks() {
  if (!window.runtime) return false;

  if (typeof window.runtime.EventsOn === 'function') {
    window.runtime.EventsOn('conversion:progress', (progress) => {
      const pct = Math.max(0, Math.min(100, Number(progress && progress.percent) || 0));
      $('progressBar').style.width = `${pct}%`;
      $('progressPercent').textContent = `${pct}%`;
      $('progressBytes').textContent = progress && progress.total > 0
        ? `${formatBytes(progress.bytes)} / ${formatBytes(progress.total)}`
        : formatBytes(progress && progress.bytes);
    });
  }

  if (typeof window.runtime.OnFileDrop === 'function') {
    window.runtime.OnFileDrop((_x, _y, paths) => {
      if (Array.isArray(paths) && paths.length > 0) inspectDropped(paths[0]);
    }, true);
  }
  return true;
}

function waitForWails(attempt = 0) {
  if (window.go && window.go.main && window.go.main.App && registerRuntimeHooks()) {
    document.documentElement.dataset.backend = 'ready';
    return;
  }
  if (attempt >= 100) {
    document.documentElement.dataset.backend = 'failed';
    showToast(new Error('BACKEND_NOT_READY'));
    return;
  }
  setTimeout(() => waitForWails(attempt + 1), 50);
}

function init() {
  applyText();
  $('dropZone').addEventListener('click', chooseFile);
  $('changeFile').addEventListener('click', chooseFile);
  $('convertButton').addEventListener('click', convertFile);
  $('revealButton').addEventListener('click', async () => {
    if (!lastOutputPath) return;
    try { await getApp().RevealFile(lastOutputPath); } catch (error) { showToast(error); }
  });
  $('againButton').addEventListener('click', () => {
    currentFile = null;
    lastOutputPath = '';
    showOnly('dropView');
  });
  waitForWails();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init, { once: true });
} else {
  init();
}
