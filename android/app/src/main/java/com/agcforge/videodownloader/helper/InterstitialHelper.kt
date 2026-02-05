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


    private var admobInterstitial: InterstitialAd? = null
    private var isAdmobLoading = false
    private var isUnityLoading = false
    private var isUnityLoaded = false
    private var startAppAd: StartAppAd? = null
    private var isStartIoLoading = false
    private var isStartIoLoaded = false

    private var lastShownTime = 0L

    init {
        startAppAd = StartAppAd(activity)
    }

    fun loadAd() {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled in config")
            return
        }

        loadAdmobInterstitial()
        loadUnityInterstitial()
        loadStartIoInterstitial()
    }

    fun showAd(
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)? = null,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)? = null,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)? = null
    ) {
        if (!AdsConfig.ENABLE_ADS) {
            Log.d(TAG, "Ads disabled")
            onAdFailed?.invoke(AdsConfig.AdsProvider.ADMOB)
            return
        }

        // Check interval
        if (!canShowAd()) {
            Log.d(TAG, "Interval not reached yet")
            onAdFailed?.invoke(AdsConfig.AdsProvider.ADMOB)
            return
        }

        // Try providers in priority order
        for (provider in AdsConfig.Rotation.INTERSTITIAL_PRIORITY) {
            when (provider) {
                AdsConfig.AdsProvider.ADMOB -> {
                    if (admobInterstitial != null) {
                        showAdmobInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.STARTIO -> {
                    if (isStartIoLoaded) {
                        showStartIoInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
            }
        }

        Log.d(TAG, "No ads available to show")
        onAdFailed?.invoke(AdsConfig.AdsProvider.ADMOB)
    }

    private fun canShowAd(): Boolean {
        val currentTime = System.currentTimeMillis()
        val intervalMillis = AdsConfig.INTERSTITIAL_INTERVAL_SECONDS * 1000L

        if (lastShownTime == 0L) return true

        return (currentTime - lastShownTime) >= intervalMillis
    }

    private fun loadAdmobInterstitial() {
        val isAdmobEnabled = AdsConfig.admobConfig.enable
        if (!isAdmobEnabled || isAdmobLoading) return

        isAdmobLoading = true

        val adRequest = AdRequest.Builder().build()

        AdsConfig.admobConfig.interstitialId?.let {
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
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        admobInterstitial?.apply {
            fullScreenContentCallback = object : FullScreenContentCallback() {
                override fun onAdShowedFullScreenContent() {
                    Log.d(TAG, "Admob interstitial shown")
                    lastShownTime = System.currentTimeMillis()
                    onAdShown?.invoke(AdsConfig.AdsProvider.ADMOB)
                }

                override fun onAdDismissedFullScreenContent() {
                    Log.d(TAG, "Admob interstitial dismissed")
                    admobInterstitial = null
                    loadAdmobInterstitial() // Reload
                    onAdClosed?.invoke(AdsConfig.AdsProvider.ADMOB)
                }

                override fun onAdFailedToShowFullScreenContent(adError: AdError) {
                    Log.e(TAG, "Admob show failed: ${adError.message}")
                    admobInterstitial = null
                    loadAdmobInterstitial()

                    // Try next provider
                    tryNextProvider(AdsConfig.AdsProvider.ADMOB, onAdShown, onAdClosed, onAdFailed)
                }
            }

            show(activity)
        } ?: run {
            tryNextProvider(AdsConfig.AdsProvider.ADMOB, onAdShown, onAdClosed, onAdFailed)
        }
    }

    private fun loadUnityInterstitial() {
        if (isUnityLoading) return
        AdsConfig.unityConfig.interstitialPlacement?.let {
            isUnityLoading = true

            UnityAds.load(
                AdsConfig.unityConfig.interstitialPlacement,
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
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        UnityAds.show(
            activity,
            AdsConfig.unityConfig.interstitialPlacement,
            object : IUnityAdsShowListener {
                override fun onUnityAdsShowStart(placementId: String) {
                    Log.d(TAG, "Unity interstitial started")
                    lastShownTime = System.currentTimeMillis()
                    onAdShown?.invoke(AdsConfig.AdsProvider.UNITY)
                }

                override fun onUnityAdsShowComplete(
                    placementId: String,
                    state: UnityAds.UnityAdsShowCompletionState
                ) {
                    Log.d(TAG, "Unity interstitial completed")
                    isUnityLoaded = false
                    loadUnityInterstitial() // Reload
                    onAdClosed?.invoke(AdsConfig.AdsProvider.UNITY)
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
                    tryNextProvider(AdsConfig.AdsProvider.UNITY, onAdShown, onAdClosed, onAdFailed)
                }

                override fun onUnityAdsShowClick(placementId: String) {
                    Log.d(TAG, "Unity interstitial clicked")
                }
            }
        )
    }
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
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        startAppAd?.showAd(object : AdDisplayListener {
            override fun adHidden(ad: Ad) {
                Log.d(TAG, "Start.io interstitial hidden")
                isStartIoLoaded = false
                loadStartIoInterstitial() // Reload
                onAdClosed?.invoke(AdsConfig.AdsProvider.STARTIO)
            }

            override fun adDisplayed(ad: Ad) {
                Log.d(TAG, "Start.io interstitial displayed")
                lastShownTime = System.currentTimeMillis()
                onAdShown?.invoke(AdsConfig.AdsProvider.STARTIO)
            }

            override fun adClicked(ad: Ad) {
                Log.d(TAG, "Start.io interstitial clicked")
            }

            override fun adNotDisplayed(ad: Ad) {
                Log.e(TAG, "Start.io not displayed")
                isStartIoLoaded = false
                loadStartIoInterstitial()

                // Try next provider
                tryNextProvider(AdsConfig.AdsProvider.STARTIO, onAdShown, onAdClosed, onAdFailed)
            }
        })
    }

    private fun tryNextProvider(
        failedProvider: AdsConfig.AdsProvider,
        onAdShown: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdClosed: ((AdsConfig.AdsProvider) -> Unit)?,
        onAdFailed: ((AdsConfig.AdsProvider) -> Unit)?
    ) {
        val currentIndex = AdsConfig.Rotation.INTERSTITIAL_PRIORITY.indexOf(failedProvider)
        val nextProviders = AdsConfig.Rotation.INTERSTITIAL_PRIORITY.drop(currentIndex + 1)

        for (provider in nextProviders) {
            when (provider) {
                AdsConfig.AdsProvider.ADMOB -> {
                    if (admobInterstitial != null) {
                        showAdmobInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.UNITY -> {
                    if (isUnityLoaded) {
                        showUnityInterstitial(onAdShown, onAdClosed, onAdFailed)
                        return
                    }
                }
                AdsConfig.AdsProvider.STARTIO -> {
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


    fun isAdReady(): Boolean {
        return admobInterstitial != null || isUnityLoaded || isStartIoLoaded
    }


    fun destroy() {
        admobInterstitial = null
        startAppAd = null
    }

}