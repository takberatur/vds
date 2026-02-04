package com.agcforge.videodownloader.helper

import android.app.Activity
import android.util.Log
import com.google.android.gms.ads.AdError
import com.google.android.gms.ads.AdRequest
import com.google.android.gms.ads.FullScreenContentCallback
import com.google.android.gms.ads.LoadAdError
import com.google.android.gms.ads.interstitial.InterstitialAd
import com.google.android.gms.ads.interstitial.InterstitialAdLoadCallback
import com.startapp.sdk.adsbase.Ad
import com.startapp.sdk.adsbase.StartAppAd
import com.startapp.sdk.adsbase.adlisteners.AdDisplayListener
import com.startapp.sdk.adsbase.adlisteners.AdEventListener
import com.unity3d.ads.IUnityAdsLoadListener
import com.unity3d.ads.IUnityAdsShowListener
import com.unity3d.ads.UnityAds
import kotlinx.coroutines.flow.first

class InterstitialHelper (private val activity: Activity) {
    private val TAG = "InterstitialHelper"

    // Admob
    private var admobInterstitial: InterstitialAd? = null
    private var isAdmobLoading = false

    // Unity Ads
    private var isUnityLoading = false
    private var isUnityLoaded = false

    // Start.io
    private var startAppAd: StartAppAd? = null
    private var isStartIoLoading = false
    private var isStartIoLoaded = false

    private var lastShownTime = 0L

    init {
        startAppAd = StartAppAd(activity)
    }

    /**
     * Load all ads providers
     */
    fun loadAd() {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled in config")
            return
        }

        loadAdmobInterstitial()
        loadUnityInterstitial()
        loadStartIoInterstitial()
    }

    /**
     * Show interstitial with auto rotation
     */
    fun showAd(
        onAdShown: ((AdsProvider) -> Unit)? = null,
        onAdClosed: ((AdsProvider) -> Unit)? = null,
        onAdFailed: ((AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            onAdFailed?.invoke(AdsProvider.ADMOB)
            return
        }

        // Check interval
        if (!canShowAd()) {
            Log.d(TAG, "Interval not reached yet")
            onAdFailed?.invoke(AdsProvider.ADMOB)
            return
        }

        // Try providers in priority order
        for (provider in AdsConfig.Rotation.getInterstitialPriority()) {
            when (provider) {
                AdsProvider.ADMOB -> {
                    if (admobInterstitial != null) {
                        showAdmobInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.STARTIO -> {
                    if (isStartIoLoaded) {
                        showStartIoInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
            }
        }

        Log.d(TAG, "No ads available to show")
        onAdFailed?.invoke(AdsProvider.ADMOB)
    }

    /**
     * Check if ad can be shown based on interval
     */
    private fun canShowAd(): Boolean {
        val currentTime = System.currentTimeMillis()
        val intervalMillis = AdsConfig.INTERSTITIAL_INTERVAL_SECONDS * 1000L

        if (lastShownTime == 0L) return true

        return (currentTime - lastShownTime) >= intervalMillis
    }

    // ========== ADMOB INTERSTITIAL ==========

    private fun loadAdmobInterstitial() {
        val isAdmobEnabled = AdsConfig.isAdmobEnabled()
        if (!isAdmobEnabled || isAdmobLoading) return

        isAdmobLoading = true

        val adRequest = AdRequest.Builder().build()

        AdsConfig.ADMOB_INTERSTITIAL_ID?.let {
            InterstitialAd.load(
                activity,
                it,
                adRequest,
                object : InterstitialAdLoadCallback() {
                    override fun onAdLoaded(interstitialAd: InterstitialAd) {
                        Log.d(TAG, "Admob interstitial loaded")
                        admobInterstitial = interstitialAd
                        isAdmobLoading = false
                    }

                    override fun onAdFailedToLoad(loadAdError: LoadAdError) {
                        Log.e(TAG, "Admob interstitial failed: ${loadAdError.message}")
                        admobInterstitial = null
                        isAdmobLoading = false
                    }
                }
            )
        }
    }

    private fun showAdmobInterstitial(
        onAdShown: ((AdsProvider) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        admobInterstitial?.apply {
            fullScreenContentCallback = object : FullScreenContentCallback() {
                override fun onAdShowedFullScreenContent() {
                    Log.d(TAG, "Admob interstitial shown")
                    lastShownTime = System.currentTimeMillis()
                    onAdShown?.invoke(AdsProvider.ADMOB)
                }

                override fun onAdDismissedFullScreenContent() {
                    Log.d(TAG, "Admob interstitial dismissed")
                    admobInterstitial = null
                    loadAdmobInterstitial() // Reload
                    onAdClosed?.invoke(AdsProvider.ADMOB)
                }

                override fun onAdFailedToShowFullScreenContent(adError: AdError) {
                    Log.e(TAG, "Admob show failed: ${adError.message}")
                    admobInterstitial = null
                    loadAdmobInterstitial()

                    // Try next provider
                    tryNextProvider(AdsProvider.ADMOB, onAdShown, onAdClosed, onAdFailed)
                }
            }

            show(activity)
        } ?: run {
            tryNextProvider(AdsProvider.ADMOB, onAdShown, onAdClosed, onAdFailed)
        }
    }

    // ========== UNITY INTERSTITIAL ==========

    private fun loadUnityInterstitial() {
        if (isUnityLoading) return
        AdsConfig.UNITY_INTERSTITIAL_ID?.let {
            isUnityLoading = true

            UnityAds.load(
                AdsConfig.UNITY_INTERSTITIAL_ID,
                object : IUnityAdsLoadListener {
                    override fun onUnityAdsAdLoaded(placementId: String) {
                        Log.d(TAG, "Unity interstitial loaded")
                        isUnityLoaded = true
                        isUnityLoading = false
                    }

                    override fun onUnityAdsFailedToLoad(
                        placementId: String,
                        error: UnityAds.UnityAdsLoadError,
                        message: String
                    ) {
                        Log.e(TAG, "Unity interstitial failed: $message")
                        isUnityLoaded = false
                        isUnityLoading = false
                    }
                }
            )
        }
    }

    private fun showUnityInterstitial(
        onAdShown: ((AdsProvider) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        UnityAds.show(
            activity,
            AdsConfig.UNITY_INTERSTITIAL_ID,
            object : IUnityAdsShowListener {
                override fun onUnityAdsShowStart(placementId: String) {
                    Log.d(TAG, "Unity interstitial started")
                    lastShownTime = System.currentTimeMillis()
                    onAdShown?.invoke(AdsProvider.UNITY)
                }

                override fun onUnityAdsShowComplete(
                    placementId: String,
                    state: UnityAds.UnityAdsShowCompletionState
                ) {
                    Log.d(TAG, "Unity interstitial completed")
                    isUnityLoaded = false
                    loadUnityInterstitial() // Reload
                    onAdClosed?.invoke(AdsProvider.UNITY)
                }

                override fun onUnityAdsShowFailure(
                    placementId: String,
                    error: UnityAds.UnityAdsShowError,
                    message: String
                ) {
                    Log.e(TAG, "Unity show failed: $message")
                    isUnityLoaded = false
                    loadUnityInterstitial()

                    // Try next provider
                    tryNextProvider(AdsProvider.UNITY, onAdShown, onAdClosed, onAdFailed)
                }

                override fun onUnityAdsShowClick(placementId: String) {
                    Log.d(TAG, "Unity interstitial clicked")
                }
            }
        )
    }

    // ========== START.IO INTERSTITIAL ==========

    private fun loadStartIoInterstitial() {
        if (isStartIoLoading) return

        isStartIoLoading = true

        startAppAd?.loadAd(object : AdEventListener {
            override fun onReceiveAd(ad: Ad) {
                Log.d(TAG, "Start.io interstitial loaded")
                isStartIoLoaded = true
                isStartIoLoading = false
            }

            override fun onFailedToReceiveAd(ad: Ad?) {
                Log.e(TAG, "Start.io interstitial failed")
                isStartIoLoaded = false
                isStartIoLoading = false
            }
        })
    }

    private fun showStartIoInterstitial(
        onAdShown: ((AdsProvider) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        startAppAd?.showAd(object : AdDisplayListener {
            override fun adHidden(ad: Ad) {
                Log.d(TAG, "Start.io interstitial hidden")
                isStartIoLoaded = false
                loadStartIoInterstitial() // Reload
                onAdClosed?.invoke(AdsProvider.STARTIO)
            }

            override fun adDisplayed(ad: Ad) {
                Log.d(TAG, "Start.io interstitial displayed")
                lastShownTime = System.currentTimeMillis()
                onAdShown?.invoke(AdsProvider.STARTIO)
            }

            override fun adClicked(ad: Ad) {
                Log.d(TAG, "Start.io interstitial clicked")
            }

            override fun adNotDisplayed(ad: Ad) {
                Log.e(TAG, "Start.io not displayed")
                isStartIoLoaded = false
                loadStartIoInterstitial()

                // Try next provider
                tryNextProvider(AdsProvider.STARTIO, onAdShown, onAdClosed, onAdFailed)
            }
        })
    }

    // ========== HELPER METHODS ==========

    private fun tryNextProvider(
        failedProvider: AdsProvider,
        onAdShown: ((AdsProvider) -> Unit)?,
        onAdClosed: ((AdsProvider) -> Unit)?,
        onAdFailed: ((AdsProvider) -> Unit)?
    ) {
        val currentIndex = AdsConfig.Rotation.getInterstitialPriority().indexOf(failedProvider)
        val nextProviders = AdsConfig.Rotation.getInterstitialPriority().drop(currentIndex + 1)

        for (provider in nextProviders) {
            when (provider) {
                AdsProvider.ADMOB -> {
                    if (admobInterstitial != null) {
                        showAdmobInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsProvider.STARTIO -> {
                    if (isStartIoLoaded) {
                        showStartIoInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
            }
        }

        // All providers failed
        Log.e(TAG, "All providers failed")
        onAdFailed?.invoke(failedProvider)
    }

    /**
     * Check if any ad is ready
     */
    fun isAdReady(): Boolean {
        return admobInterstitial != null || isUnityLoaded || isStartIoLoaded
    }

    /**
     * Destroy ads
     */
    fun destroy() {
        admobInterstitial = null
        startAppAd = null
    }

}