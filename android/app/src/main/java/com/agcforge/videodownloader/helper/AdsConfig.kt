package com.agcforge.videodownloader.helper

import android.content.Context
import android.util.Log
import com.agcforge.videodownloader.data.model.Application
import com.agcforge.videodownloader.utils.PreferenceManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

object AdsConfig {

    // Static config properties (backward compatibility)
    var TEST_MODE: Boolean = true
    var INTERSTITIAL_INTERVAL_SECONDS: Int = 60
    var ENABLE_ADS: Boolean = true

    var ENABLE_ADMOB: Boolean = false
    var ENABLE_UNITY: Boolean = false
    var ENABLE_STARTIO: Boolean = false

    // Ad Unit IDs
    var ADMOB_BANNER_ID: String? = null
    var ADMOB_INTERSTITIAL_ID: String? = null
    var ADMOB_REWARDED_ID: String? = null
    var ADMOB_NATIVE_ID: String? = null

    var UNITY_GAME_ID: String? = null
    var UNITY_BANNER_ID: String? = null
    var UNITY_INTERSTITIAL_ID: String? = null
    var UNITY_REWARDED_ID: String? = null
    var UNITY_NATIVE_ID: String? = null

    var STARTIO_APP_ID: String? = null

    var ONESIGNAL_ID: String? = null

    private var preferenceManager: PreferenceManager? = null
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    fun initialize(context: Context) {
        preferenceManager = PreferenceManager(context)
        loadConfigFromPreferences()
    }

    private fun loadConfigFromPreferences() {
        preferenceManager?.let { manager ->
            TEST_MODE = manager.getBooleanSync("ads_test_mode") ?: true
            INTERSTITIAL_INTERVAL_SECONDS = manager.getIntSync("ads_interstitial_interval") ?: 60
            ENABLE_ADS = manager.getBooleanSync("ads_enable") ?: true

            // Load application config
            val appConfig = manager.getApplicationSync()
            updateFromApplicationConfig(appConfig)
        }
    }

    fun updateFromApplicationConfig(appConfig: Application?) {
        appConfig?.let { app ->
            // Update monetization flag
            ENABLE_ADS = app.enableMonetization

            // Update each provider flags
            ENABLE_ADMOB = app.enableMonetization && app.enableAdmob
            ENABLE_UNITY = app.enableMonetization && app.enableUnityAd
            ENABLE_STARTIO = app.enableMonetization && app.enableStartApp

            // Update Admob IDs
            ADMOB_BANNER_ID = if (ENABLE_ADMOB) app.admobBannerAdUnitId else null
            ADMOB_INTERSTITIAL_ID = if (ENABLE_ADMOB) app.admobInterstitialAdUnitId else null
            ADMOB_REWARDED_ID = if (ENABLE_ADMOB) app.admobRewardedAdUnitId else null
            ADMOB_NATIVE_ID = if (ENABLE_ADMOB) app.admobNativeAdUnitId else null

            // Update Unity IDs
            UNITY_GAME_ID = if (ENABLE_UNITY) app.unityAdUnitId else null
            UNITY_BANNER_ID = if (ENABLE_UNITY) app.unityBannerAdUnitId else null
            UNITY_INTERSTITIAL_ID = if (ENABLE_UNITY) app.unityInterstitialAdUnitId else null
            UNITY_REWARDED_ID = if (ENABLE_UNITY) app.unityRewardedAdUnitId else null
            UNITY_NATIVE_ID = if (ENABLE_UNITY) app.unityNativeAdUnitId else null

            // Update Start.io ID
            STARTIO_APP_ID = if (ENABLE_STARTIO) app.startAppAdUnitId else null

            // Save to preferences for persistence
            scope.launch {
                preferenceManager?.saveBoolean("ads_enable", ENABLE_ADS)
            }

            Log.d("AdsConfig", "Updated config: Admob=$ENABLE_ADMOB, Unity=$ENABLE_UNITY, Start.io=$ENABLE_STARTIO")
        }
    }

    // Helper functions
    fun isAdmobEnabled(): Boolean = ENABLE_ADS && ENABLE_ADMOB
    fun isUnityEnabled(): Boolean = ENABLE_ADS && ENABLE_UNITY
    fun isStartIoEnabled(): Boolean = ENABLE_ADS && ENABLE_STARTIO

    fun getEnabledProviders(): List<AdsProvider> {
        val providers = mutableListOf<AdsProvider>()
        if (isAdmobEnabled()) providers.add(AdsProvider.ADMOB)
        if (isUnityEnabled()) providers.add(AdsProvider.UNITY)
        if (isStartIoEnabled()) providers.add(AdsProvider.STARTIO)
        return providers
    }

    // Setter functions
    fun setTestMode(testMode: Boolean) {
        TEST_MODE = testMode
        scope.launch {
            preferenceManager?.saveBoolean("ads_test_mode", testMode)
        }
    }

    fun setInterstitialIntervalSeconds(seconds: Int) {
        INTERSTITIAL_INTERVAL_SECONDS = seconds
        scope.launch {
            preferenceManager?.saveInt("ads_interstitial_interval", seconds)
        }
    }

    fun setEnableAds(enable: Boolean) {
        ENABLE_ADS = enable
        scope.launch {
            preferenceManager?.saveBoolean("ads_enable", enable)
        }
    }

    // Rotation logic
    object Rotation {
        fun getInterstitialPriority(): List<AdsProvider> {
            return getPriorityList(AdType.INTERSTITIAL)
        }

        fun getRewardPriority(): List<AdsProvider> {
            return getPriorityList(AdType.REWARDED)
        }

        fun getBannerPriority(): List<AdsProvider> {
            return getPriorityList(AdType.BANNER)
        }

        fun getNativePriority(): List<AdsProvider> {
            return getPriorityList(AdType.NATIVE)
        }

        private fun getPriorityList(adType: AdType): List<AdsProvider> {
            val providers = mutableListOf<AdsProvider>()

            if (isAdmobEnabled() && hasAdUnit(AdsProvider.ADMOB, adType)) {
                providers.add(AdsProvider.ADMOB)
            }

            if (isUnityEnabled() && hasAdUnit(AdsProvider.UNITY, adType)) {
                providers.add(AdsProvider.UNITY)
            }

            if (isStartIoEnabled() && hasAdUnit(AdsProvider.STARTIO, adType)) {
                providers.add(AdsProvider.STARTIO)
            }

            return if (providers.isEmpty()) {
                // Fallback
                when (adType) {
                    AdType.NATIVE -> listOf(AdsProvider.ADMOB)
                    else -> listOf(AdsProvider.ADMOB, AdsProvider.UNITY, AdsProvider.STARTIO)
                }
            } else {
                providers
            }
        }

        private fun hasAdUnit(provider: AdsProvider, adType: AdType): Boolean {
            return when (provider) {
                AdsProvider.ADMOB -> when (adType) {
                    AdType.BANNER -> !ADMOB_BANNER_ID.isNullOrEmpty()
                    AdType.INTERSTITIAL -> !ADMOB_INTERSTITIAL_ID.isNullOrEmpty()
                    AdType.NATIVE -> !ADMOB_NATIVE_ID.isNullOrEmpty()
                    AdType.REWARDED -> !ADMOB_REWARDED_ID.isNullOrEmpty()
                }
                AdsProvider.UNITY -> when (adType) {
                    AdType.BANNER -> !UNITY_BANNER_ID.isNullOrEmpty()
                    AdType.INTERSTITIAL -> !UNITY_INTERSTITIAL_ID.isNullOrEmpty()
                    AdType.NATIVE -> !UNITY_NATIVE_ID.isNullOrEmpty()
                    AdType.REWARDED -> !UNITY_REWARDED_ID.isNullOrEmpty()
                }
                AdsProvider.STARTIO -> !STARTIO_APP_ID.isNullOrEmpty()
            }
        }
    }


    enum class AdType {
        BANNER,
        INTERSTITIAL,
        NATIVE,
        REWARDED
    }
}
