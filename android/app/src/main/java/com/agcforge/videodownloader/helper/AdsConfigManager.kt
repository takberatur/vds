package com.agcforge.videodownloader.helper

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.agcforge.videodownloader.data.api.VideoDownloaderRepository
import com.agcforge.videodownloader.data.model.Application
import com.agcforge.videodownloader.utils.PreferenceManager
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class AdsConfigManager(
    private val preferenceManager: PreferenceManager,
    private val repository: VideoDownloaderRepository
) : ViewModel() {

    private val _adsConfig = MutableStateFlow(AdsRuntimeConfig.getDefault())
    val adsConfig: StateFlow<AdsRuntimeConfig> = _adsConfig.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    init {
        loadInitialConfig()
        observeApplicationConfig()
    }

    private fun loadInitialConfig() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                val appConfig = preferenceManager.getApplication()
                updateAdsConfig(appConfig)
            } finally {
                _isLoading.value = false
            }
        }
    }

    private fun observeApplicationConfig() {
        viewModelScope.launch {
            preferenceManager.applicationConfig.collect { appConfig ->
                updateAdsConfig(appConfig)
            }
        }
    }

    private fun updateAdsConfig(appConfig: Application?) {
        appConfig?.let { app ->
            val newConfig = AdsRuntimeConfig(
                testMode = AdsConfig.TEST_MODE,
                enableMonetization = app.enableMonetization,
                interstitialInterval = AdsConfig.INTERSTITIAL_INTERVAL_SECONDS,
                enableAdmob = app.enableMonetization && app.enableAdmob,
                enableUnity = app.enableMonetization && app.enableUnityAd,
                enableStartIo = app.enableMonetization && app.enableStartApp,
                admobIds = AdMobIds(
                    banner = if (app.enableAdmob) app.admobBannerAdUnitId else null,
                    interstitial = if (app.enableAdmob) app.admobInterstitialAdUnitId else null,
                    native = if (app.enableAdmob) app.admobNativeAdUnitId else null,
                    rewarded = if (app.enableAdmob) app.admobRewardedAdUnitId else null
                ),
                unityIds = UnityIds(
                    banner = if (app.enableUnityAd) app.unityBannerAdUnitId else null,
                    interstitial = if (app.enableUnityAd) app.unityInterstitialAdUnitId else null,
                    native = if (app.enableUnityAd) app.unityNativeAdUnitId else null,
                    rewarded = if (app.enableUnityAd) app.unityRewardedAdUnitId else null
                ),
                startIoId = if (app.enableStartApp) app.startAppAdUnitId else null,
                lastUpdated = System.currentTimeMillis()
            )

            _adsConfig.value = newConfig

            // Update static AdsConfig for backward compatibility
            AdsConfig.updateFromApplicationConfig(app)
        }
    }

    fun refreshConfig() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                repository.getApplication()
                    .onSuccess { app ->
                        preferenceManager.saveApplication(app)
                        updateAdsConfig(app)
                    }
                    .onFailure {
                        // Handle error silently or show toast
                    }
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun setTestMode(testMode: Boolean) {
        AdsConfig.setTestMode(testMode)
        _adsConfig.value = _adsConfig.value.copy(testMode = testMode)
    }

    fun setInterstitialInterval(seconds: Int) {
        AdsConfig.setInterstitialIntervalSeconds(seconds)
        _adsConfig.value = _adsConfig.value.copy(interstitialInterval = seconds)
    }

    fun setEnableAds(enable: Boolean) {
        AdsConfig.setEnableAds(enable)
        _adsConfig.value = _adsConfig.value.copy(enableMonetization = enable)
    }
}

data class AdsRuntimeConfig(
    val testMode: Boolean,
    val enableMonetization: Boolean,
    val interstitialInterval: Int,
    val enableAdmob: Boolean,
    val enableUnity: Boolean,
    val enableStartIo: Boolean,
    val admobIds: AdMobIds,
    val unityIds: UnityIds,
    val startIoId: String?,
    val lastUpdated: Long
) {
    companion object {
        fun getDefault(): AdsRuntimeConfig {
            return AdsRuntimeConfig(
                testMode = true,
                enableMonetization = true,
                interstitialInterval = 60,
                enableAdmob = true,
                enableUnity = false,
                enableStartIo = false,
                admobIds = AdMobIds(null, null, null, null),
                unityIds = UnityIds(null, null, null, null),
                startIoId = null,
                lastUpdated = 0
            )
        }
    }

    fun isProviderEnabled(provider: AdsProvider): Boolean {
        return when (provider) {
            AdsProvider.ADMOB -> enableAdmob
            AdsProvider.UNITY -> enableUnity
            AdsProvider.STARTIO -> enableStartIo
        }
    }
}

data class AdMobIds(
    val banner: String?,
    val interstitial: String?,
    val native: String?,
    val rewarded: String?
)

data class UnityIds(
    val banner: String?,
    val interstitial: String?,
    val native: String?,
    val rewarded: String?
)

enum class AdsProvider {
    ADMOB,
    UNITY,
    STARTIO
}