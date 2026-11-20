package webui

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/More678/strip-comments/internal/glossary"
	"github.com/More678/strip-comments/internal/lang"
	"github.com/More678/strip-comments/internal/scan"
	"github.com/More678/strip-comments/internal/textio"
	"github.com/More678/strip-comments/internal/udiff"
	"github.com/More678/strip-comments/internal/verify"
	"github.com/More678/strip-comments/internal/walk"
)

const maxSource = 256 << 10

// A non-nil empty source is valid: stripping a comments-only file can also give an empty result.
type previewRequest struct {
	Source       *string `json:"source"`
	Filename     string  `json:"filename"`
	Language     string  `json:"language"`
	KeepDoc      bool    `json:"keepDoc"`
	KeepBlank    bool    `json:"keepBlank"`
	KeepLicense  bool    `json:"keepLicense"`
	Replace      bool    `json:"replace"`
	CommentsOnly bool    `json:"commentsOnly"`
	KeepComments bool    `json:"keepComments"`
}

type previewResult struct {
	Language     string `json:"language"`
	DetectedBy   string `json:"detectedBy"`
	Output       string `json:"output"`
	Diff         string `json:"diff"`
	DiffOmitted  bool   `json:"diffOmitted"`
	Changed      bool   `json:"changed"`
	Comments     int    `json:"comments"`
	RemovedLines int    `json:"removedLines"`
	Terms        int    `json:"terms"`
}

func preview(w http.ResponseWriter, r *http.Request, table *glossary.Table) {
	media, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" || (params["charset"] != "" && !strings.EqualFold(params["charset"], "utf-8")) {
		problem(w, http.StatusUnsupportedMediaType, "预览需要 UTF-8 JSON。")
		return
	}
	// JSON escaping can use up to six wire bytes per source byte.
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 6*maxSource+4096))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			problem(w, http.StatusRequestEntityTooLarge, "请求过大，源码上限为 256 KiB。")
		} else {
			problem(w, http.StatusBadRequest, "无法读取预览请求。")
		}
		return
	}
	var in previewRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if !utf8.Valid(body) || dec.Decode(&in) != nil || dec.Decode(new(any)) != io.EOF || in.Source == nil {
		problem(w, http.StatusBadRequest, "预览请求必须包含 UTF-8 源码和有效的选项。")
		return
	}
	if len(*in.Source) > maxSource {
		problem(w, http.StatusRequestEntityTooLarge, "源码超过 256 KiB。")
		return
	}
	if strings.IndexByte(*in.Source, 0) >= 0 {
		problem(w, http.StatusBadRequest, "只接受 UTF-8 文本，不接受二进制文件。")
		return
	}
	if len(in.Filename) > 255 || strings.ContainsAny(in.Filename, `/\:`) || strings.IndexFunc(in.Filename, unicode.IsControl) != -1 || in.Filename == "." || in.Filename == ".." {
		problem(w, http.StatusBadRequest, "只接受文件名，不接受路径。")
		return
	}
	if in.Filename != "" && walk.SecretReason(in.Filename) != "" {
		problem(w, http.StatusForbidden, "已跳过可能包含凭据的文件。")
		return
	}
	l, by := lang.ByName(in.Language), "manual"
	if in.Language == "auto" || in.Language == "" {
		l, by = lang.Detect(in.Filename, []byte(*in.Source))
		if l != nil && l.Risky {
			problem(w, http.StatusBadRequest, "这种语言使用启发式规则，请手动选择后预览。")
			return
		}
	}
	if l == nil {
		problem(w, http.StatusBadRequest, "无法确定语言，请手动选择后再生成预览。")
		return
	}
	if !in.Replace && (in.CommentsOnly || in.KeepComments) {
		problem(w, http.StatusBadRequest, "保留全部注释和仅替换注释需要先启用审计术语替换。")
		return
	}
	result, err := transform(in, l, table)
	if err != nil {
		problem(w, http.StatusUnprocessableEntity, "无法安全预览："+err.Error())
		return
	}
	result.DetectedBy = by
	writeJSON(w, http.StatusOK, result)
}

func transform(in previewRequest, l *lang.Language, table *glossary.Table) (previewResult, error) {
	opt := scan.Options{KeepDoc: in.KeepDoc, KeepBlank: in.KeepBlank, KeepLicense: in.KeepLicense,
		TermsInComments: in.CommentsOnly, KeepComments: in.KeepComments}
	if in.Replace {
		opt.Glossary = table
	}
	text, format := textio.Decode([]byte(*in.Source))
	sr := scan.Strip(text, l, opt)
	if sr.Unsafe {
		return previewResult{}, errors.New(sr.Reason)
	}
	out := previewResult{Language: l.Name, Output: *in.Source, Comments: sr.Comments, RemovedLines: sr.RemovedLines, Terms: len(sr.Terms)}
	if bytes.Equal(text, sr.Output) {
		return out, nil // do not normalize mixed line endings on unchanged input
	}
	if err := verify.Check(text, sr.Output, l); err != nil {
		return previewResult{}, err
	}
	out.Output, out.Changed = string(textio.Encode(sr.Output, format)), true
	// The CLI's line diff uses quadratic memory; bound its table for HTTP requests.
	lines := bytes.Count(text, []byte{'\n'}) + 1
	afterLines := bytes.Count(sr.Output, []byte{'\n'}) + 1
	if int64(lines)*int64(afterLines) > 1_000_000 || max(lines, afterLines) > 5000 {
		out.DiffOmitted = true
	} else {
		name := in.Filename
		if name == "" {
			name = l.Name
		}
		out.Diff = udiff.Unified(name, text, sr.Output)
	}
	return out, nil
}
