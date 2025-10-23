import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost, getApiUrl } from "../shared/utils";
import "./TravelDiary.css";

interface DiaryEntry {
    id: number;
    date: string;
    title: string;
    content: string;
    images: string[];
    location?: string;
    mood?: string;
}

const TravelDiary: React.FC = () => {
    const navigate = useNavigate();
    const [entries, setEntries] = useState<DiaryEntry[]>([]);
    const [title, setTitle] = useState("");
    const [content, setContent] = useState("");
    const [location, setLocation] = useState("");
    const [mood, setMood] = useState("");
    const [images, setImages] = useState<string[]>([]);
    const [editingId, setEditingId] = useState<number | null>(null);
    const [filter, setFilter] = useState<string>("all");
    const [viewingEntry, setViewingEntry] = useState<DiaryEntry | null>(null);
    const [loading, setLoading] = useState(false);

    // 从后端加载日记
    useEffect(() => {
        loadDiaries();
    }, []);

    const loadDiaries = async () => {
        try {
            setLoading(true);
            const response = await apiGet("/api/diaries");
            if (response.success && response.data) {
                setEntries(response.data);
            }
        } catch (error) {
            console.error("加载日记失败", error);
        } finally {
            setLoading(false);
        }
    };

    const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = e.target.files;
        if (!files) return;

        Array.from(files).forEach((file) => {
            if (file.size > 2 * 1024 * 1024) {
                alert("图片大小不能超过2MB");
                return;
            }

            const reader = new FileReader();
            reader.onload = (event) => {
                const result = event.target?.result as string;
                setImages((prev) => [...prev, result]);
            };
            reader.readAsDataURL(file);
        });
    };

    const handleRemoveImage = (index: number) => {
        setImages((prev) => prev.filter((_, i) => i !== index));
    };

    const handleAddOrUpdateEntry = async () => {
        if (!title.trim() && !content.trim()) {
            alert("请至少填写标题或内容");
            return;
        }

        try {
            setLoading(true);
            if (editingId) {
                // 更新现有日记
                const response = await fetch(getApiUrl(`/api/diaries/${editingId}`), {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${localStorage.getItem("token")}`,
                    },
                    body: JSON.stringify({
                        date: new Date().toLocaleString("zh-CN"),
                        title,
                        content,
                        location,
                        mood,
                        images,
                    }),
                });

                const result = await response.json();
                if (result.success) {
                    await loadDiaries();
                    setEditingId(null);
                } else {
                    alert("更新失败");
                }
            } else {
                // 添加新日记
                const response = await apiPost("/api/diaries", {
                    date: new Date().toLocaleString("zh-CN"),
                    title,
                    content,
                    location,
                    mood,
                    images,
                });

                if (response.success) {
                    await loadDiaries();
                } else {
                    alert("添加失败");
                }
            }

            // 重置表单
            resetForm();
        } catch (error) {
            console.error("操作失败", error);
            alert("操作失败，请重试");
        } finally {
            setLoading(false);
        }
    };

    const resetForm = () => {
        setTitle("");
        setContent("");
        setLocation("");
        setMood("");
        setImages([]);
        setEditingId(null);
    };

    const handleEdit = (entry: DiaryEntry) => {
        setTitle(entry.title);
        setContent(entry.content);
        setLocation(entry.location || "");
        setMood(entry.mood || "");
        setImages(entry.images);
        setEditingId(entry.id);
        window.scrollTo({ top: 0, behavior: "smooth" });
    };

    const handleDelete = async (id: number) => {
        if (!confirm("确定要删除这条日记吗？")) {
            return;
        }

        try {
            setLoading(true);
            const response = await fetch(getApiUrl(`/api/diaries/${id}`), {
                method: "DELETE",
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("token")}`,
                },
            });

            const result = await response.json();
            if (result.success) {
                await loadDiaries();
            } else {
                alert("删除失败");
            }
        } catch (error) {
            console.error("删除失败", error);
            alert("删除失败，请重试");
        } finally {
            setLoading(false);
        }
    };

    const filteredEntries = entries.filter((entry) => {
        if (filter === "all") return true;
        return entry.mood === filter;
    });

    const moodEmojis: { [key: string]: string } = {
        happy: "😊",
        excited: "🤩",
        relaxed: "😌",
        tired: "😴",
        amazed: "😲",
    };

    // 如果正在查看某条日记，显示详情页
    if (viewingEntry) {
        return (
            <div className="diary-detail-page">
                <div className="detail-header">
                    <button onClick={() => setViewingEntry(null)} className="back-btn">
                        ← 返回列表
                    </button>
                    <div className="detail-actions">
                        <button onClick={() => {
                            handleEdit(viewingEntry);
                            setViewingEntry(null);
                        }} className="edit-action-btn">
                            ✏️ 编辑
                        </button>
                        <button onClick={() => {
                            handleDelete(viewingEntry.id);
                            setViewingEntry(null);
                        }} className="delete-action-btn">
                            🗑️ 删除
                        </button>
                    </div>
                </div>

                <div className="detail-content">
                    <div className="detail-title-row">
                        {viewingEntry.mood && (
                            <span className="detail-mood">{moodEmojis[viewingEntry.mood]}</span>
                        )}
                        <h1>{viewingEntry.title || "无标题"}</h1>
                    </div>

                    <div className="detail-meta">
                        <span className="detail-date">📅 {viewingEntry.date}</span>
                        {viewingEntry.location && (
                            <span className="detail-location">📍 {viewingEntry.location}</span>
                        )}
                    </div>

                    <div className="detail-text">{viewingEntry.content}</div>

                    {viewingEntry.images.length > 0 && (
                        <div className="detail-images">
                            {viewingEntry.images.map((img, index) => (
                                <img
                                    key={index}
                                    src={img}
                                    alt={`${viewingEntry.title} - ${index + 1}`}
                                    className="detail-image"
                                />
                            ))}
                        </div>
                    )}
                </div>
            </div>
        );
    }

    return (
        <div className="travel-diary">
            <div className="diary-header">
                <button onClick={() => navigate(-1)} className="back-btn">← 返回</button>
                <h2>✈️ 旅行日记</h2>
            </div>

            <div className="diary-editor">
                <h3>{editingId ? "编辑日记" : "写新日记"}</h3>

                <input
                    type="text"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    placeholder="日记标题..."
                    className="diary-title-input"
                />

                <div className="date-location-row">
                    <input
                        type="date"
                        className="diary-date-input"
                    />
                    <input
                        type="text"
                        value={location}
                        onChange={(e) => setLocation(e.target.value)}
                        placeholder="地点（可选）"
                        className="diary-location-input"
                    />
                </div>

                <div className="mood-selector">
                    <label>心情：</label>
                    {Object.entries(moodEmojis).map(([key, emoji]) => (
                        <button
                            key={key}
                            className={`mood-btn ${mood === key ? "active" : ""}`}
                            onClick={() => setMood(key)}
                            type="button"
                        >
                            {emoji}
                        </button>
                    ))}
                </div>

                <textarea
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    placeholder="记录你的旅行点滴..."
                    rows={6}
                    className="diary-content-input"
                />

                <div className="image-upload-section">
                    <label htmlFor="image-upload" className="upload-label">
                        📷 添加图片
                    </label>
                    <input
                        id="image-upload"
                        type="file"
                        accept="image/*"
                        multiple
                        onChange={handleImageUpload}
                        style={{ display: "none" }}
                    />
                    <div className="image-preview-container">
                        {images.map((img, index) => (
                            <div key={index} className="image-preview">
                                <img src={img} alt={`预览 ${index + 1}`} />
                                <button
                                    className="remove-image-btn"
                                    onClick={() => handleRemoveImage(index)}
                                >
                                    ×
                                </button>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="editor-actions">
                    <button
                        onClick={handleAddOrUpdateEntry}
                        className="save-btn"
                        disabled={loading}
                    >
                        {loading ? "保存中..." : (editingId ? "💾 保存修改" : "➕ 添加日记")}
                    </button>
                    {editingId && (
                        <button onClick={resetForm} className="cancel-btn">
                            取消编辑
                        </button>
                    )}
                </div>
            </div>

            <div className="diary-filter">
                <button
                    className={filter === "all" ? "active" : ""}
                    onClick={() => setFilter("all")}
                >
                    全部 ({entries.length})
                </button>
                {Object.entries(moodEmojis).map(([key, emoji]) => {
                    const count = entries.filter((e) => e.mood === key).length;
                    if (count === 0) return null;
                    return (
                        <button
                            key={key}
                            className={filter === key ? "active" : ""}
                            onClick={() => setFilter(key)}
                        >
                            {emoji} ({count})
                        </button>
                    );
                })}
            </div>

            <ul className="diary-list">
                {loading && entries.length === 0 ? (
                    <div className="empty-state">
                        <p>⏳ 加载中...</p>
                    </div>
                ) : filteredEntries.length === 0 ? (
                    <div className="empty-state">
                        <p>📝 还没有日记，开始记录你的旅程吧！</p>
                    </div>
                ) : (
                    filteredEntries.map((entry) => (
                        <li
                            key={entry.id}
                            className="diary-entry"
                            onClick={() => setViewingEntry(entry)}
                            style={{ cursor: 'pointer' }}
                        >
                            <div className="entry-header">
                                <div className="entry-title-row">
                                    {entry.mood && (
                                        <span className="entry-mood">{moodEmojis[entry.mood]}</span>
                                    )}
                                    <h4>{entry.title || "无标题"}</h4>
                                </div>
                                <div className="entry-actions" onClick={(e) => e.stopPropagation()}>
                                    <button onClick={() => handleEdit(entry)} className="edit-btn">
                                        ✏️
                                    </button>
                                    <button onClick={() => handleDelete(entry.id)} className="delete-btn">
                                        🗑️
                                    </button>
                                </div>
                            </div>
                            <div className="diary-meta">
                                <span className="diary-date">📅 {entry.date}</span>
                                {entry.location && (
                                    <span className="diary-location">📍 {entry.location}</span>
                                )}
                            </div>
                            <div className="diary-content">{entry.content}</div>
                            {entry.images.length > 0 && (
                                <div className="diary-images">
                                    {entry.images.slice(0, 3).map((img, index) => (
                                        <img
                                            key={index}
                                            src={img}
                                            alt={`${entry.title} - ${index + 1}`}
                                            className="diary-image"
                                        />
                                    ))}
                                    {entry.images.length > 3 && (
                                        <div className="more-images">+{entry.images.length - 3}</div>
                                    )}
                                </div>
                            )}
                        </li>
                    ))
                )}
            </ul>
        </div>
    );
};

export default TravelDiary;
