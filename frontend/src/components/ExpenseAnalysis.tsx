import { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import './ExpenseAnalysis.css';

interface AnalysisData {
    analysis: string;
    query?: string;
    category?: string;
    from?: string;
    to?: string;
}

export default function ExpenseAnalysis() {
    const location = useLocation();
    const navigate = useNavigate();
    const [analysisData, setAnalysisData] = useState<AnalysisData | null>(null);

    useEffect(() => {
        const state = location.state as { analysis?: AnalysisData };
        if (state?.analysis) {
            setAnalysisData(state.analysis);
        }
    }, [location]);

    return (
        <div className="expense-analysis-page">
            <div className="analysis-container">
                <div className="analysis-header">
                    <button className="back-button" onClick={() => navigate('/budget')}>
                        ← 返回预算管理
                    </button>
                    <h2>💡 AI 消费分析</h2>
                </div>

                {analysisData ? (
                    <div className="analysis-content">
                        {(analysisData.query || analysisData.category || analysisData.from) && (
                            <div className="analysis-filters">
                                <h3>分析条件</h3>
                                {analysisData.query && (
                                    <div className="filter-item">
                                        <span className="filter-label">查询：</span>
                                        <span className="filter-value">{analysisData.query}</span>
                                    </div>
                                )}
                                {analysisData.category && (
                                    <div className="filter-item">
                                        <span className="filter-label">分类：</span>
                                        <span className="filter-value">{analysisData.category}</span>
                                    </div>
                                )}
                                {analysisData.from && analysisData.to && (
                                    <div className="filter-item">
                                        <span className="filter-label">时间范围：</span>
                                        <span className="filter-value">{analysisData.from} 至 {analysisData.to}</span>
                                    </div>
                                )}
                            </div>
                        )}

                        <div className="analysis-result">
                            <h3>分析结果</h3>
                            <div className="analysis-text">
                                {analysisData.analysis.split('\n').map((line, index) => (
                                    <p key={index}>{line}</p>
                                ))}
                            </div>
                        </div>
                    </div>
                ) : (
                    <div className="no-analysis">
                        <p>暂无分析数据</p>
                        <button className="back-button" onClick={() => navigate('/budget')}>
                            返回预算管理
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
