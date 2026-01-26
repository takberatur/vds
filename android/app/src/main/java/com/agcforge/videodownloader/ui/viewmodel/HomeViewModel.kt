package com.agcforge.videodownloader.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.agcforge.videodownloader.data.api.VideoDownloaderRepository
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.data.model.Platform
import com.agcforge.videodownloader.data.websocket.CentrifugoEvent
import com.agcforge.videodownloader.data.websocket.CentrifugoManager
import com.agcforge.videodownloader.data.websocket.DownloadTaskEvent
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.Resource
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class HomeViewModel(application: Application) : AndroidViewModel(application) {

    private val repository = VideoDownloaderRepository()
    private val centrifugoManager = CentrifugoManager.getInstance(application)
    private val preferenceManager = PreferenceManager(application)

    private val _platforms = MutableStateFlow<Resource<List<Platform>>>(Resource.Loading())
    val platforms: StateFlow<Resource<List<Platform>>> = _platforms.asStateFlow()

    private val _downloadResult = MutableStateFlow<Resource<DownloadTask>>(Resource.Loading())
    val downloadResult: StateFlow<Resource<DownloadTask>> = _downloadResult.asStateFlow()

    // WebSocket connection state
    private val _wsConnectionState = MutableStateFlow<CentrifugoEvent>(CentrifugoEvent.Disconnected)
    val wsConnectionState: StateFlow<CentrifugoEvent> = _wsConnectionState.asStateFlow()

    // Real-time download events
    private val _realtimeDownloadEvent = MutableStateFlow<DownloadTaskEvent?>(null)
    val realtimeDownloadEvent: StateFlow<DownloadTaskEvent?> = _realtimeDownloadEvent.asStateFlow()

    init {
        observeWebSocketEvents()
        initializeWebSocket()
    }

    private fun initializeWebSocket() {
        viewModelScope.launch {
            // Get user ID from preferences
            preferenceManager.userId.collect { userId ->
                if (!userId.isNullOrEmpty()) {
                    // Get auth token
                    preferenceManager.authToken.collect { token ->
                        centrifugoManager.initialize(userId, token)
                        centrifugoManager.connect()
                    }
                }
            }
        }
    }

    private fun observeWebSocketEvents() {
        viewModelScope.launch {
            centrifugoManager.connectionState.collect { state ->
                _wsConnectionState.value = state
            }
        }

        viewModelScope.launch {
            centrifugoManager.downloadEvents.collect { event ->
                event?.let { _realtimeDownloadEvent.value = it }
            }
        }
    }

    fun loadPlatforms() {
        viewModelScope.launch {
            _platforms.value = Resource.Loading()

            repository.getPlatforms()
                .onSuccess { platformList ->
                    _platforms.value = Resource.Success(platformList)
                }
                .onFailure { error ->
                    _platforms.value = Resource.Error(error.message ?: "Failed to load platforms")
                }
        }
    }

    fun createDownload(url: String, platformId: String? = null) {
        viewModelScope.launch {
            _downloadResult.value = Resource.Loading()

            repository.createDownload(url, platformId)
                .onSuccess { task ->
                    _downloadResult.value = Resource.Success(task)

                    // Subscribe to download channel untuk real-time updates
                    centrifugoManager.subscribeToDownloadChannel(task.id)
                }
                .onFailure { error ->
                    _downloadResult.value = Resource.Error(error.message ?: "Download failed")
                }
        }
    }

    fun connectWebSocket() {
        centrifugoManager.connect()
    }

    fun disconnectWebSocket() {
        centrifugoManager.disconnect()
    }

    fun subscribeToPublicDownloads() {
        centrifugoManager.subscribeToPublicDownloads()
    }
}


