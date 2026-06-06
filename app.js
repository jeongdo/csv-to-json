const LANG = {
    ko: {
        title: "CSV to JSON 변환기",
        notice1: "※ 첫 번째 행(Header)에 반드시 칼럼명이 포함되어 있어야 합니다.",
        notice2: "※ 첫 행의 칼럼명을 기반으로 JSON의 key 값이 자동으로 생성됩니다.",
        upload: "업로드 및 JSON 다운로드",
        converting: "변환 중...",
        success: "변환이 완료되었습니다.",
        error: "변환 중 오류가 발생했습니다.",

        errors: {
            INVALID_METHOD: "잘못된 접근입니다.",
            UPLOAD_FAILED: "업로드 파일이 너무 크거나 손상되었습니다.",
            FILE_READ_FAILED: "파일을 읽는 중 오류가 발생했습니다.",
            HEADER_READ_FAILED: "헤더를 읽을 수 없습니다.",
            EMPTY_HEADER: "빈 헤더가 존재합니다.",
            DUPLICATE_HEADER: function (detail) {
                return `중복 헤더 발견: ${detail}`;
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
        success: "Conversion completed.",
        error: "An error occurred during conversion.",

        errors: {
            INVALID_METHOD: "Invalid request.",
            UPLOAD_FAILED: "File is too large or corrupted.",
            FILE_READ_FAILED: "Failed to read file.",
            HEADER_READ_FAILED: "Failed to read CSV header.",
            EMPTY_HEADER: "Empty header detected.",
            DUPLICATE_HEADER: function (detail) {
                return `Duplicate header: ${detail}`
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

document
    .getElementById("convertForm")
    .addEventListener("submit", async function (e) {

        e.preventDefault();

        const button = document.getElementById("convertButton");

        button.disabled = true;
        button.innerText = TEXT.converting;

        const fileInput = this.querySelector('input[name="csvFile"]');

        const file = fileInput.files[0];

        let filename = "result.json";

        if (file) {
            const originalName = file.name;
            const lastDot = originalName.lastIndexOf(".");

            filename =
                (lastDot !== -1
                    ? originalName.substring(0, lastDot)
                    : originalName)
                + ".json";
        }

        const formData = new FormData(this);

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
                    getErrorMessage(
                        errorData.code,
                        errorData.detail
                    ),
                    "error"
                );

                button.disabled = false;
                button.innerText = TEXT.upload;
                return;
            }

            const blob = await response.blob();

            const url = window.URL.createObjectURL(blob);

            const a = document.createElement("a");

            a.href = url;
            a.download = filename;

            document.body.appendChild(a);

            a.click();

            a.remove();

            window.URL.revokeObjectURL(url);

            showToast(TEXT.success, "success");

        } catch (err) {
            showToast(TEXT.error, "error");
        }

        button.disabled = false;
        button.innerText = TEXT.upload;
    });

window.addEventListener(
    "beforeunload",
    function () {
        navigator.sendBeacon("/shutdown");
    }
);


const dropZone = document.getElementById("dropZone");
const fileInput = document.getElementById("csvFile");
const selectedFileName = document.getElementById("selectedFileName");

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

            fileInput.files = files;

            selectedFileName.innerText =
                files[0].name;
        }
    }
);

fileInput.addEventListener(
    "change",
    function () {

        if (this.files.length > 0) {

            selectedFileName.innerText =
                this.files[0].name;
        }
    }
);