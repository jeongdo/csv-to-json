const LANG = {
    ko: {
        title: "CSV to JSON 변환기",
        notice1: "※ 첫 번째 행(Header)에 반드시 칼럼명이 포함되어 있어야 합니다.",
        notice2: "※ 첫 행의 칼럼명을 기반으로 JSON의 key 값이 자동으로 생성됩니다.",
        upload: "업로드 및 JSON 다운로드",
        converting: "변환 중...",
        success: "✓ 다운로드 완료",
        error: "변환 중 오류가 발생했습니다.",
        noFile: "파일을 먼저 선택해 주세요.",
        dropText: "CSV 파일을 여기에 드래그하거나",
        noFileSelected: "선택된 파일 없음",

        errors: {
            INVALID_METHOD: "잘못된 접근입니다.",
            UPLOAD_FAILED: "업로드 파일이 너무 크거나 손상되었습니다.",
            FILE_READ_FAILED: "파일을 읽는 중 오류가 발생했습니다.",
            HEADER_READ_FAILED: "헤더를 읽을 수 없습니다.",
            EMPTY_HEADER: "빈 헤더가 존재합니다.",
            DUPLICATE_HEADER: function (detail) {
                const parts = detail.split(',');
                const firstName = parts[0];
                const otherCount = parseInt(parts[1], 10);
                if (otherCount > 0) {
                    return `중복 헤더 발견: ${firstName} 외 ${otherCount} 건`;
                } else {
                    return `중복 헤더 발견: ${firstName}`;
                }
            },
            MIXED_DELIMITER_DETECTED: "CSV 구분자가 혼합되어 있어 처리할 수 없습니다.",
            CSV_PARSE_FAILED: "CSV 형식이 올바르지 않습니다.",
            UNKNOWN_ERROR: "알 수 없는 오류가 발생했습니다."
        }
    },

    en: {
        title: "CSV to JSON Converter",
        notice1: "First row must contain column headers.",
        notice2: "JSON keys are generated automatically from the header row.",
        upload: "Upload and Download JSON",
        converting: "Converting...",
        success: "✓ Download complete",
        error: "An error occurred during conversion.",
        noFile: "Please select a file first.",
        dropText: "Drag & Drop CSV File Here",
        noFileSelected: "No file selected",

        errors: {
            INVALID_METHOD: "Invalid request.",
            UPLOAD_FAILED: "File is too large or corrupted.",
            FILE_READ_FAILED: "Failed to read file.",
            HEADER_READ_FAILED: "Failed to read CSV header.",
            EMPTY_HEADER: "Empty header detected.",
            DUPLICATE_HEADER: function (detail) {
                return `Duplicate header: ${detail}`
            },
            DUPLICATE_HEADER: function (detail) {
                const parts = detail.split(',');
                const firstName = parts[0];
                const otherCount = parseInt(parts[1], 10);
                if (otherCount > 0) {
                    return `Duplicate header: ${firstName} and ${otherCount} others`;
                } else {
                    return `Duplicate header: ${firstName}`;
                }
            },
            MIXED_DELIMITER_DETECTED: "The CSV file cannot be processed because it contains mixed delimiters.",
            CSV_PARSE_FAILED: "Invalid CSV format.",
            UNKNOWN_ERROR: "An unknown error occurred."
        }
    }
};

const browserLang = navigator.language.toLowerCase();
const TEXT = browserLang.startsWith("ko") ? LANG.ko : LANG.en;

function applyLanguage() {
    document.title = TEXT.title;
    document.getElementById("title").innerText = TEXT.title;
    document.getElementById("notice1").innerText = TEXT.notice1;
    document.getElementById("notice2").innerText = TEXT.notice2;
    document.getElementById("convertButton").innerText = TEXT.upload;
}

function forceResize() {
    if (window.resizeTo) {
        window.resizeTo(600, 450);
    }
}

window.addEventListener(
    "DOMContentLoaded",
    function () {
        applyLanguage();
        forceResize();
    }
);

let toastTimer = null;

function showToast(message, type) {
    const toast = document.getElementById("toast");
    toast.textContent = message;
    toast.className = type + " show";

    if (toastTimer) {
        clearTimeout(toastTimer);
    }

    toast.onclick = function () {
        toast.classList.remove("show");
        if (toastTimer) {
            clearTimeout(toastTimer);
        }
    };

    toastTimer = setTimeout(function () {
        toast.classList.remove("show");
    }, 3000);
}

function getErrorMessage(code, detail) {
    const errorEntry = TEXT.errors[code];

    if (!errorEntry) {
        console.warn("Unknown error code:", code, detail);
        return TEXT.errors.UNKNOWN_ERROR;
    }

    if (typeof errorEntry === "function") {
        return errorEntry(detail);
    }

    return errorEntry;
}

// [FIX 1] 파일 상태를 별도 변수로 관리 — fileInput.files 직접 의존 제거
let selectedFile = null;

function setSelectedFile(file) {
    selectedFile = file;
    const selectedFileName = document.getElementById("selectedFileName");
    if (file) {
        selectedFileName.innerText = file.name;
    } else {
        selectedFileName.innerText = TEXT.noFileSelected;
    }
}

document
    .getElementById("convertForm")
    .addEventListener("submit", async function (e) {

        e.preventDefault();

        // [FIX 1] selectedFile 변수로 판단
        if (!selectedFile) {
            showToast(TEXT.noFile, "error");
            return;
        }

        const button = document.getElementById("convertButton");
        button.disabled = true;
        button.innerText = TEXT.converting;

        const originalName = selectedFile.name;
        const lastDot = originalName.lastIndexOf(".");
        const filename =
            (lastDot !== -1
                ? originalName.substring(0, lastDot)
                : originalName)
            + ".json";

        // [FIX 1] FormData를 직접 구성해서 selectedFile을 명시적으로 첨부
        const formData = new FormData();
        formData.append("csvFile", selectedFile);

        try {
            const response = await fetch(
                "/convert",
                {
                    method: "POST",
                    body: formData
                }
            );

            if (!response.ok) {
                const errorData = await response.json();
                showToast(
                    getErrorMessage(errorData.code, errorData.detail),
                    "error"
                );
                button.disabled = false;
                button.innerText = TEXT.upload;
                return;
            }

            const blob = await response.blob();

            if (window.showSaveFilePicker) {
                // [FIX 2] 저장 대화상자를 직접 띄워서 완료 시점을 정확히 감지
                try {
                    const fileHandle = await window.showSaveFilePicker({
                        suggestedName: filename,
                        types: [{
                            description: "JSON File",
                            accept: { "application/json": [".json"] }
                        }]
                    });
                    const writable = await fileHandle.createWritable();
                    await writable.write(blob);
                    await writable.close();
                    showToast(TEXT.success, "success");
                } catch (saveErr) {
                    // 사용자가 저장 대화상자를 취소한 경우 — 에러 아님, 조용히 복귀
                    if (saveErr.name !== "AbortError") {
                        showToast(TEXT.error, "error");
                    }
                }
            } else {
                // showSaveFilePicker 미지원 브라우저 fallback
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = filename;
                document.body.appendChild(a);
                a.click();
                a.remove();
                window.URL.revokeObjectURL(url);
                showToast(TEXT.success, "success");
            }

        } catch (err) {
            showToast(TEXT.error, "error");
        }

        button.disabled = false;
        button.innerText = TEXT.upload;
    });

const _shutdownToken = new URLSearchParams(window.location.search).get("token") ?? "";

window.addEventListener(
    "beforeunload",
    function () {
        navigator.sendBeacon("/shutdown?token=" + _shutdownToken);
    }
);

const dropZone = document.getElementById("dropZone");
const fileInput = document.getElementById("csvFile");

dropZone.addEventListener(
    "dragover",
    function (e) {
        e.preventDefault();
        dropZone.classList.add("drag-over");
    }
);

dropZone.addEventListener(
    "dragleave",
    function () {
        dropZone.classList.remove("drag-over");
    }
);

dropZone.addEventListener(
    "drop",
    function (e) {
        e.preventDefault();
        dropZone.classList.remove("drag-over");

        const files = e.dataTransfer.files;
        if (files.length > 0) {
            // [FIX 1] selectedFile 변수에 저장
            setSelectedFile(files[0]);
        }
    }
);

fileInput.addEventListener(
    "change",
    function () {
        if (this.files.length > 0) {
            // [FIX 1] 파일 선택 시 selectedFile 업데이트
            setSelectedFile(this.files[0]);
        } else {
            // [FIX 1] 선택 취소 시 명시적으로 null 처리
            setSelectedFile(null);
        }
        // input 값을 리셋해도 selectedFile은 유지되므로 다음 드랍과 충돌 없음
        this.value = "";
    }
);
