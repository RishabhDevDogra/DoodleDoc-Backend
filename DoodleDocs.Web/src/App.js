import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { API_URL, WS_HUB_URL } from './config';
import './App.css';
import DocumentList from './components/DocumentList';
import DocumentEditor from './components/DocumentEditor';
import VersionHistory from './components/VersionHistory';
import TopNavbar from './components/TopNavbar';
import ShareModal from './components/ShareModal';
import Comments from './components/Comments';
import ErrorBoundary from './components/ErrorBoundary';
import { useToast } from './components/Toast';
import { getOrCreateUserId } from './utils/userSession';

function App() {
  const { addToast } = useToast();
  const { docId: urlDocId } = useParams();
  const navigate = useNavigate();
  const [documents, setDocuments] = useState([]);
  const [selectedDocId, setSelectedDocId] = useState(urlDocId || null);
  const [selectedDoc, setSelectedDoc] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [userName, setUserName] = useState('');
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  const [showComments, setShowComments] = useState(false);
  const isEditingRef = useRef(false);
  const titleSaveTimerRef = useRef(null);

  // Initialize user session on mount
  useEffect(() => {
    const { userName: uname } = getOrCreateUserId();
    setUserName(uname);
  }, []);

  // Set up WebSocket connection for real-time updates
  useEffect(() => {
    let active = true;
    const ws = new WebSocket(WS_HUB_URL);

    ws.onopen = () => {
      if (active) console.log('WebSocket Connected');
    };

    ws.onmessage = (event) => {
      if (!active) return;
      try {
        const { type, payload } = JSON.parse(event.data);

        if (type === 'DocumentCreated') {
          console.log('Document created:', payload.documentId, payload.title);
          fetchDocuments();
        } else if (type === 'DocumentUpdated') {
          console.log('Document updated:', payload.documentId);
          fetchDocuments();
          if (selectedDocId === payload.documentId && !isEditingRef.current) {
            fetchDocument(payload.documentId);
          }
        } else if (type === 'DocumentDeleted') {
          console.log('Document deleted:', payload.documentId);
          fetchDocuments();
          if (selectedDocId === payload.documentId) {
            setSelectedDocId(null);
            setSelectedDoc(null);
          }
        }
      } catch (err) {
        console.error('WebSocket message parse error:', err);
      }
    };

    ws.onerror = () => {
      if (active) console.error('WebSocket error');
    };

    return () => {
      active = false;
      ws.close();
    };
  }, [selectedDocId]);

  // Fetch all documents on mount
  useEffect(() => {
    fetchDocuments();
  }, []);

  // Keep URL in sync with selected document
  useEffect(() => {
    if (selectedDocId) {
      navigate(`/${selectedDocId}`, { replace: true });
    }
  }, [selectedDocId]);

  // Auto-create first document if none exist and select it
  useEffect(() => {
    if (documents.length === 0 && !selectedDocId) {
      createNewDocument();
    } else if (documents.length > 0 && !selectedDocId) {
      // If we have documents but none selected, select the first one
      setSelectedDocId(documents[0].id);
    }
  }, [documents, selectedDocId]);

  // Fetch document details when selected
  useEffect(() => {
    if (selectedDocId) {
      // Fetch document without waiting (background task)
      fetchDocument(selectedDocId);
    }
  }, [selectedDocId]);

  const fetchDocuments = async () => {
    try {
      const res = await fetch(`${API_URL}/api/document`);
      const data = await res.json();
      setDocuments(data);
    } catch (err) {
      console.error('Error fetching documents:', err);
    }
  };

  const fetchDocument = async (id) => {
    try {
      const res = await fetch(`${API_URL}/api/document/${id}`);
      const data = await res.json();
      setSelectedDoc(data);
    } catch (err) {
      console.error('Error fetching document:', err);
    }
  };

  const createNewDocument = async () => {
    try {
      const { userId, userName: uname } = getOrCreateUserId();
      const res = await fetch(`${API_URL}/api/document`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: 'Untitled Doodle', userId, userName: uname })
      });
      const newDoc = await res.json();
      setDocuments([newDoc, ...documents]);
      setSelectedDocId(newDoc.id);
    } catch (err) {
      console.error('Error creating document:', err);
    }
  };

  const updateDocument = async (id, title, content) => {
    try {
      const { userId, userName: uname } = getOrCreateUserId();
      const res = await fetch(`${API_URL}/api/document/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, userId, userName: uname })
      });
      const updated = await res.json();
      // Only update state if we're not actively editing
      if (!isEditingRef.current) {
        setSelectedDoc(updated);
      }
      setDocuments(documents.map(d => d.id === id ? updated : d));
    } catch (err) {
      console.error('Error updating document:', err);
    }
  };

  const handleTitleChange = (newTitle) => {
    if (selectedDoc) {
      isEditingRef.current = true;
      setSelectedDoc({ ...selectedDoc, title: newTitle });
      
      // Clear previous timer
      if (titleSaveTimerRef.current) {
        clearTimeout(titleSaveTimerRef.current);
      }
      
      // Debounce the actual save
      titleSaveTimerRef.current = setTimeout(async () => {
        await updateDocument(selectedDoc.id, newTitle, selectedDoc.content);
        isEditingRef.current = false;
      }, 500);
    }
  };

  const deleteDocument = async (id) => {
    try {
      await fetch(`${API_URL}/api/document/${id}`, { method: 'DELETE' });
      setDocuments(documents.filter(d => d.id !== id));
      if (selectedDocId === id) {
        setSelectedDocId(null);
        setSelectedDoc(null);
      }
    } catch (err) {
      console.error('Error deleting document:', err);
    }
  };

  const duplicateDocument = async (docToDupe) => {
    try {
      const { userId, userName: uname } = getOrCreateUserId();
      const res = await fetch(`${API_URL}/api/document`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: `${docToDupe.title} (Copy)`, userId, userName: uname })
      });
      const newDoc = await res.json();
      // Copy content from original
      await fetch(`${API_URL}/api/document/${newDoc.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: `${docToDupe.title} (Copy)`, content: docToDupe.content, userId, userName: uname })
      });
      setDocuments([newDoc, ...documents]);
      setSelectedDocId(newDoc.id);
    } catch (err) {
      console.error('Error duplicating document:', err);
    }
  };

  return (
    <ErrorBoundary>
      <div className="app-wrapper">
        <TopNavbar 
          userName={userName} 
          documentTitle={selectedDoc?.title ?? ''}
          onTitleChange={handleTitleChange}
          onShare={() => setIsShareModalOpen(true)}
          onNewDoodle={createNewDocument}
        />
        <div className="app">
          <div className="editor-area">
            {selectedDoc ? (
              <DocumentEditor
                document={selectedDoc}
                onUpdate={updateDocument}
                onToggleHistory={() => setShowVersionHistory(!showVersionHistory)}
                showHistory={showVersionHistory}
                onToggleComments={() => setShowComments(!showComments)}
                showComments={showComments}
              />
            ) : (
              <div className="no-doc">
                <p>Loading your masterpiece...</p>
              </div>
            )}
          </div>
          {selectedDoc && showVersionHistory && (
            <div className="version-history-panel">
              <VersionHistory documentId={selectedDoc.id} userName={userName} onClose={() => setShowVersionHistory(false)} />
            </div>
          )}
          {selectedDoc && showComments && (
            <div className="comments-panel-wrapper">
              <Comments documentId={selectedDoc.id} userName={userName} isOpen={showComments} onClose={() => setShowComments(false)} />
            </div>
          )}
        </div>
        {selectedDoc && (
          <ShareModal 
            documentId={selectedDoc.id}
            documentTitle={selectedDoc.title}
            isOpen={isShareModalOpen}
            onClose={() => setIsShareModalOpen(false)}
          />
        )}
      </div>
    </ErrorBoundary>
  );
}

export default App;
